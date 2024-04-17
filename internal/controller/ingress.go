package controller

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	"reflect"

	aloystechv1 "aloys.tech/api/v1"
	"aloys.tech/internal/utils"
	netv1 "k8s.io/api/networking/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *AppReconciler) reconcileIngress(ctx context.Context, app *aloystechv1.App) (ctrl.Result, error) {
	ingressName := app.Name + "-ingress"
	logger := log.FromContext(ctx).WithName("reconcileIngress").WithName(ingressName)
	// 创建使用模版，是为了可以在模块添加一些亲和性，资源请求这些配置,这步骤在前面是想判断一下ingress的内容是否需要更新
	appIngress := utils.NewIngress(app)
	if err := ctrl.SetControllerReference(app, appIngress, r.Scheme); err != nil {
		logger.Error(err, "Failed to set the controller reference for the app ingress,will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	ing := &netv1.Ingress{}
	err := r.Get(ctx, GetNamespacedName(app.Name, "-ingress", app.Namespace), ing)
	// ingress 存在，并且开启 并且nodeport 没设置
	if err == nil {
		logger.Info("The Ingress already exists.")
		// 这里就要额外添加判断了，svc的判断存在不存在没有必要
		// 额外加2个条件，1.ingress要是开启状态
		// 					2.svc不能是nodePort，不然就要删除ingress (这里使用webhook实现更方便)
		// 判断是否需要更新ingress,
		if app.Spec.Ingress.IsEnable == true && app.Spec.Service.NodePort == 0 {
			if !reflect.DeepEqual(ing.Spec, appIngress.Spec) {
				logger.Info("This Ingress has been updated. Update it. ")
				if err := r.Update(ctx, appIngress); err != nil {
					logger.Error(err, "Failed to update the Ingress,will requeue after a short time.")
					return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
				}
				logger.Info("The Ingress updated successfully.")
			}
			// 判断ingress status 是否需要更新
			if !reflect.DeepEqual(ing.Status, app.Status.IngressSpec) {
				logger.Info("This Ingress Status has been updated. Update it.")
				app.Status.IngressSpec = ing.Spec
				// 更新Status
				if err := r.Status().Update(ctx, app); err != nil {
					logger.Error(err, "Failed to update the app status,will requeue after a short time.")
					return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
				}
				logger.Info("The Ingress status has been updated successfully.")
			}
			return ctrl.Result{}, nil
		}
		// ingress 存在，但是没开启，或者nodeport 设置了
		logger.Info("The ingress already exists, delete ingress.")
		if err := r.Delete(ctx, appIngress); err != nil {
			logger.Error(err, "Failed to delete the Ingress,will requeue after a short time.")
			return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
		}
		logger.Info("The ingress deleted successfully.")
		return ctrl.Result{}, nil
	}
	// 错误是NotFound 直接结束本轮
	if !errors.IsNotFound(err) {
		logger.Error(err, "Failed to get the Ingress,will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	// 创建
	// 额外加2个条件，1.ingress要是开启状态
	// 					2.svc不能是nodePort，否则不创建
	if app.Spec.Ingress.IsEnable == true && app.Spec.Service.NodePort == 0 {
		logger.Info("The ingress start creating.")
		if err := r.Create(ctx, appIngress); err != nil {
			logger.Error(err, "Failed to create the Ingress,will requeue after a short time.")
			return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
		}
		logger.Info("The ingress has been created.")
		return ctrl.Result{}, nil
	}
	if app.Spec.Ingress.IsEnable == true && app.Spec.Service.NodePort != 0 {
		logger.Info("Both Service and Ingress are set, and Service takes effect.")
		return ctrl.Result{}, nil
	}
	// ingress不存在，也不需要创建，直接返回就行
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, err
}
