package controller

import (
	"context"
	"reflect"

	aloystechv1 "aloys.tech/api/v1"
	"aloys.tech/internal/utils"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *AppReconciler) reconcileHorizontalPodAutoscaler(ctx context.Context, app *aloystechv1.App) (ctrl.Result, error) {
	hpaName := app.Name + "-hpa"
	logger := log.FromContext(ctx).WithName("reconcileHorizontalPodAutoscaler").WithName(hpaName)
	// 创建使用模版，是为了可以在模块添加一些亲和性，资源请求这些配置,这步骤在前面是想判断一下deploy的内容是否需要更新
	appHPA := utils.NewHorizontalPodAutoscaler(app)
	if err := ctrl.SetControllerReference(app, appHPA, r.Scheme); err != nil {
		logger.Error(err, "Failed to set the controller reference for the app HPA,will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	hpa := &autoscalingv2.HorizontalPodAutoscaler{}
	hpaErr := r.Get(ctx, GetNamespacedName(app.Name, "-hpa", app.Namespace), hpa)
	// dp := &appsv1.Deployment{}
	// deployErr := r.Get(ctx, GetNamespacedName(app.Name, "-deploy", app.Namespace), dp)
	if hpaErr == nil {
		logger.Info("The HPA already exists.")
		// hpa存在，要判断deploy是否存在,这个逻辑走不到，因为deploy删除后会直接创建
		// deploy不存在，就要删除hpa
		// 删除这个判断逻辑，监听的时候，如果deplo不存在马上创建，就不存在deploy不存在的情况这个代码就执行不到
		// if errors.IsNotFound(deployErr) {
		// 	logger.Info("The Deployment is not found.", "deployment", dp.Name)
		// 	err := r.Delete(ctx, hpa)
		// 	if err != nil {
		// 		logger.Error(err, "Failed to delete the HPA,will requeue after a short time.")
		// 		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
		// 	}
		// 	logger.Info("The HPA has been deleted.")
		// 	return ctrl.Result{}, nil
		// }
		// 这时候认为hpa和deploy都存在
		// 更新hpa
		if !reflect.DeepEqual(hpa.Spec, appHPA.Spec) {
			logger.Info("This HPA has been updated. Update it.")
			err := r.Update(ctx, appHPA)
			if err != nil {
				logger.Error(err, "Failed to update the HPA ,will requeue after a short time.")
				return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
			}
			logger.Info("The HPA updated successfully.")
		}
		// 更新hpa status
		if !reflect.DeepEqual(hpa.Status, appHPA.Status) {
			logger.Info("This HPA Status has been updated. Update it.")
			app.Status.HorizontalPodAutoscalerStatus = hpa.Status
			if err := r.Status().Update(ctx, app); err != nil {
				logger.Error(err, "Failed to update the HPA,will requeue after a short time.")
				return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
			}
			logger.Info("The HPA status updated successfully.")
		}
		return ctrl.Result{}, nil
	}
	// 如果hpa不是不存在的错误，并且deployErr 不是不存在
	// if !errors.IsNotFound(hpaErr) && deployErr != nil {
	// 	logger.Error(hpaErr, "Failed to get the HPA,will requeue after a short time.")
	// 	return ctrl.Result{RequeueAfter: GenericRequeueDuration}, hpaErr
	// }
	if !errors.IsNotFound(hpaErr) {
		logger.Error(hpaErr, "Failed to get the HPA,will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, hpaErr
	}
	// 其他就是返回hpa错误
	logger.Info("The HPA start creating.")
	// 创建hpa
	if err := r.Create(ctx, appHPA); err != nil {
		logger.Error(err, "Failed to create the HPA,will requeue after a short time.")
		return ctrl.Result{RequeueAfter: GenericRequeueDuration}, err
	}
	logger.Info("The HPA has been created.")
	return ctrl.Result{}, nil

}
