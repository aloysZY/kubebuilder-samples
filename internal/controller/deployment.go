package controller

import (
	"context"
	"reflect"

	aloystechv1 "aloys.tech/api/v1"
	"aloys.tech/internal/utils"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *AppReconciler) reconcileDeployment(ctx context.Context, app *aloystechv1.App) (ctrl.Result, error) {
	deployName := app.Name + "-deploy"
	logger := log.FromContext(ctx).WithName("reconcileDeployment").WithName(deployName)
	// 创建使用模版，是为了可以在模块添加一些亲和性，资源请求这些配置,这步骤在前面是想判断一下deploy的内容是否需要更新
	appDeploy := utils.NewDeployment(app)
	if err := ctrl.SetControllerReference(app, appDeploy, r.Scheme); err != nil {
		logger.Error(err, "Failed to set the controller reference for the app deployment,will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	dp := &appsv1.Deployment{}
	err := r.Get(ctx, GetNamespacedName(app.Name, "-deploy", app.Namespace), dp)
	// 能查询到 err == nil
	if err == nil {
		logger.Info("The Deployment already exists.")
		// 判断是否需要更新deploy,
		if !reflect.DeepEqual(dp.Spec, appDeploy.Spec) {
			// 不相同进行更新,这里想比较一样，看谁的版本谁最新的，进行更新那个版本
			// 这里是想如果直接修改了deploy，那就按照修改的执行，但是不能确认修改了什么，监听的时候就很杂乱
			// if dp.GetResourceVersion() > appDeploy.GetResourceVersion() {
			// 	appDeploy.Spec.Replicas = dp.Spec.Replicas
			// }
			logger.Info("This Deployment has been updated. Update it. ")
			if err := r.Update(ctx, appDeploy); err != nil {
				logger.Error(err, "Failed to update the deployment,will requeue after a short time.")
				r.Eventer.Eventf(appDeploy, "Normal", "DeploymentUpdated", "Failed to update the %s deployment ,will requeue after a short time. namespace: %s", appDeploy.Name, appDeploy.Namespace)
				return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
			}
			logger.Info("The Deployment updated successfully.")
			r.Eventer.Eventf(appDeploy, "Normal", "Deployment Updated", "The %s Deployment updated successfully. namespace:%s", appDeploy.Name, appDeploy.Namespace)
		}
		// 判断deploy status 是否需要更新
		if !reflect.DeepEqual(dp.Status, app.Status.DeploymentStatus) {
			logger.Info("This Deployment Status has been updated. Update it.")
			app.Status.DeploymentStatus = dp.Status
			// 更新Status
			if err := r.Status().Update(ctx, app); err != nil {
				logger.Error(err, "Failed to update the app status,will requeue after a short time.")
				r.Eventer.Eventf(appDeploy, "Normal", "App status Updated", "Failed to update the %s status,will requeue after a short time. namespace: %s", app.Name, app.Namespace)
				return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
			}
			r.Eventer.Eventf(appDeploy, "Normal", "Deployment  status Updated", "This Deployment status has been updated.")
			logger.Info("The Deployment status has been updated successfully.")
		}
		return ctrl.Result{}, nil
	}
	// 错误不是NotFound 直接结束本轮
	if !errors.IsNotFound(err) {
		logger.Error(err, "Failed to get the Deployment,will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	// 创建
	// logger.Info("The Deployment start creating.")
	if err := r.Create(ctx, appDeploy); err != nil {
		logger.Error(err, "Failed to create the deployment,will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	r.Eventer.Eventf(appDeploy, "Normal", "Deployment created", "This Deployment created.")
	logger.Info("The Deployment has been created.")
	return ctrl.Result{}, nil
}
