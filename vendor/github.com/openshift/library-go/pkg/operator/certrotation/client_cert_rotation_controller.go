package certrotation

import (
	"context"
	"fmt"
	"time"

	operatorv1 "github.com/openshift/api/operator/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/condition"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/v1helpers"
)

const (
	// CertificateNotBeforeAnnotation contains the certificate expiration date in RFC3339 format.
	CertificateNotBeforeAnnotation = "auth.openshift.io/certificate-not-before"
	// CertificateNotAfterAnnotation contains the certificate expiration date in RFC3339 format.
	CertificateNotAfterAnnotation = "auth.openshift.io/certificate-not-after"
	// CertificateIssuer contains the common name of the certificate that signed another certificate.
	CertificateIssuer = "auth.openshift.io/certificate-issuer"
	// CertificateHostnames contains the hostnames used by a signer.
	CertificateHostnames = "auth.openshift.io/certificate-hostnames"
	// RunOnceContextKey is a context value key that can be used to call the controller Sync() and make it only run the syncWorker once and report error.
	RunOnceContextKey = "cert-rotation-controller.openshift.io/run-once"
)

// StatusReporter knows how to report the status of cert rotation
type StatusReporter interface {
	Report(ctx context.Context, syncErr error) (updated bool, updateErr error)
}

var _ StatusReporter = (*StaticPodConditionStatusReporter)(nil)

type StaticPodConditionStatusReporter struct {
	// conditionName is used in operator conditions to identify this controller, compare CertRotationDegradedConditionTypeFmt.
	conditionName string

	// Plumbing:
	OperatorClient v1helpers.StaticPodOperatorClient
}

func (s *StaticPodConditionStatusReporter) Report(ctx context.Context, syncErr error) (bool, error) {
	newCondition := operatorv1.OperatorCondition{
		Type:   fmt.Sprintf(condition.CertRotationDegradedConditionTypeFmt, s.conditionName),
		Status: operatorv1.ConditionFalse,
	}
	if syncErr != nil {
		newCondition.Status = operatorv1.ConditionTrue
		newCondition.Reason = "RotationError"
		newCondition.Message = syncErr.Error()
	}
	_, updated, updateErr := v1helpers.UpdateStaticPodStatus(ctx, s.OperatorClient, v1helpers.UpdateStaticPodConditionFn(newCondition))
	return updated, updateErr
}

// CertRotationController does:
//
// 1) continuously create a self-signed signing CA (via RotatedSigningCASecret) and store it in a secret.
// 2) maintain a CA bundle ConfigMap with all not yet expired CA certs.
// 3) continuously create a target cert and key signed by the latest signing CA and store it in a secret.
type CertRotationController struct {
	// rotatedSigningCASecret rotates a self-signed signing CA stored in a secret.
	rotatedSigningCASecret RotatedSigningCASecret
	// CABundleConfigMap maintains a CA bundle config map, by adding new CA certs coming from rotatedSigningCASecret, and by removing expired old ones.
	CABundleConfigMap CABundleConfigMap
	// RotatedSelfSignedCertKeySecret rotates a key and cert signed by a signing CA and stores it in a secret.
	RotatedSelfSignedCertKeySecret RotatedSelfSignedCertKeySecret

	// Plumbing:
	StatusReporter StatusReporter
}

func NewCertRotationController(
	name string,
	rotatedSigningCASecret RotatedSigningCASecret,
	caBundleConfigMap CABundleConfigMap,
	rotatedSelfSignedCertKeySecret RotatedSelfSignedCertKeySecret,
	recorder events.Recorder,
	reporter StatusReporter,
) factory.Controller {
	c := &CertRotationController{
		rotatedSigningCASecret:         rotatedSigningCASecret,
		CABundleConfigMap:              caBundleConfigMap,
		RotatedSelfSignedCertKeySecret: rotatedSelfSignedCertKeySecret,
		StatusReporter:                 reporter,
	}
	return factory.New().
		ResyncEvery(time.Minute).
		WithSync(c.Sync).
		WithInformers(
			rotatedSigningCASecret.Informer.Informer(),
			caBundleConfigMap.Informer.Informer(),
			rotatedSelfSignedCertKeySecret.Informer.Informer(),
		).
		WithPostStartHooks(
			c.targetCertRecheckerPostRunHook,
		).
		ToController("CertRotationController", recorder.WithComponentSuffix("cert-rotation-controller").WithComponentSuffix(name))
}

func (c CertRotationController) Sync(ctx context.Context, syncCtx factory.SyncContext) error {
	syncErr := c.syncWorker(ctx)

	// running this function with RunOnceContextKey value context will make this "run-once" without updating status.
	isRunOnce, ok := ctx.Value(RunOnceContextKey).(bool)
	if ok && isRunOnce {
		return syncErr
	}

	updated, updateErr := c.StatusReporter.Report(ctx, syncErr)
	if updateErr != nil {
		return updateErr
	}
	if updated && syncErr != nil {
		syncCtx.Recorder().Warningf("RotationError", syncErr.Error())
	}

	return syncErr
}

func (c CertRotationController) syncWorker(ctx context.Context) error {
	signingCertKeyPair, err := c.rotatedSigningCASecret.ensureSigningCertKeyPair(ctx)
	if err != nil {
		return err
	}

	cabundleCerts, err := c.CABundleConfigMap.ensureConfigMapCABundle(ctx, signingCertKeyPair)
	if err != nil {
		return err
	}

	if err := c.RotatedSelfSignedCertKeySecret.ensureTargetCertKeyPair(ctx, signingCertKeyPair, cabundleCerts); err != nil {
		return err
	}

	return nil
}

func (c CertRotationController) targetCertRecheckerPostRunHook(ctx context.Context, syncCtx factory.SyncContext) error {
	// If we have a need to force rechecking the cert, use this channel to do it.
	refresher, ok := c.RotatedSelfSignedCertKeySecret.CertCreator.(TargetCertRechecker)
	if !ok {
		return nil
	}
	targetRefresh := refresher.RecheckChannel()
	go wait.Until(func() {
		for {
			select {
			case <-targetRefresh:
				syncCtx.Queue().Add(factory.DefaultQueueKey)
			case <-ctx.Done():
				return
			}
		}
	}, time.Minute, ctx.Done())

	<-ctx.Done()
	return nil
}
