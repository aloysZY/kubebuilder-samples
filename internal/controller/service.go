package controller

import (
	"context"

	aloystechv1 "aloys.tech/api/v1"
	"aloys.tech/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *AppReconciler) reconcileService(ctx context.Context, app *aloystechv1.App) (ctrl.Result, error) {

	svcName := app.Name + "-svc"
	logger := log.FromContext(ctx).WithName("reconcileService").WithName(svcName)

	svc := &corev1.Service{}
	err := r.Get(ctx, GetNamespacedName(app.Name, "-svc", app.Namespace), svc)
	if err == nil {

	}
	if !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
	appService := utils.NewService(app)
	if err := ctrl.SetControllerReference(app, appService, r.Scheme); err != nil {
		logger.Error(err, "Failed to set the controller reference for the app appService ,will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	r.Create(ctx, appService)
	return ctrl.Result{}, nil
}
