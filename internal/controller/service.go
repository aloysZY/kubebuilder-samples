package controller

import (
	"context"
	"reflect"

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
	appService := utils.NewService(app)
	if err := ctrl.SetControllerReference(app, appService, r.Scheme); err != nil {
		logger.Error(err, "Failed to set the controller reference for the app appService ,will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	svc := &corev1.Service{}
	err := r.Get(ctx, GetNamespacedName(app.Name, "-svc", app.Namespace), svc)
	if err == nil {
		// 所以这里也不判断其他在不在，这个svc也是必须监听存在的资源，不存在马上创建
		if !reflect.DeepEqual(svc.Spec, appService.Spec) {
			// 不相同进行更新,这里想比较一样，看谁的版本谁最新的，进行更新那个版本
			// 这里是想如果直接修改了deploy，那就按照修改的执行，但是不能确认修改了什么，监听的时候就很杂乱
			// if dp.GetResourceVersion() > appDeploy.GetResourceVersion() {
			// 	appDeploy.Spec.Replicas = dp.Spec.Replicas
			// }
			logger.Info("The Service has been updated. Update it. ")
			if err := r.Update(ctx, appService); err != nil {
				logger.Error(err, "Failed to update the Service,will requeue after a short time.")
				return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
			}
			logger.Info("The Service updated successfully.")

			logger.Info("This Service Status has been updated. Update it.")
			app.Status.ServiceSpec = svc.Spec
			// 更新Status
			if err := r.Status().Update(ctx, app); err != nil {
				logger.Error(err, "Failed to update the app status,will requeue after a short time.")
				return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
			}
			logger.Info("The Service status has been updated successfully.")
		}
		return ctrl.Result{}, nil
	}
	if !errors.IsNotFound(err) {
		logger.Error(err, "Failed to get the Service, will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	logger.Info("The Service start creating.")
	if err := r.Create(ctx, appService); err != nil {
		logger.Error(err, "Failed to create the Service,will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	logger.Info("The Service has been created.")
	return ctrl.Result{}, nil
}
