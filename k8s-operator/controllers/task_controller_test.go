package controllers

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task controller", func() {
	Context("Task controller test", func() {

		const TaskName = "test-task"

		ctx := context.Background()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      TaskName,
				Namespace: TaskName,
			},
		}

		typedNamespaceName := types.NamespacedName{Name: TaskName, Namespace: TaskName}

		BeforeEach(func() {
			By("Creating the Namespace to perform the tests")
			err := k8sClient.Create(ctx, namespace)
			Expect(err).To(Not(HaveOccurred()))
		})
		AfterEach(func() {
			By("Deleting the Namespace to perform the tests")
			_ = k8sClient.Delete(ctx, namespace)
		})

		It("should successfully reconcile a Task", func() {
			By("Creating the custom resource for Kind Task")
			task := &neosyncdevv1alpha1.Task{}
			err := k8sClient.Get(ctx, typedNamespaceName, task)
			if err != nil && errors.IsNotFound(err) {
				task = &neosyncdevv1alpha1.Task{
					ObjectMeta: metav1.ObjectMeta{
						Name:      TaskName,
						Namespace: namespace.Name,
					},
					Spec: neosyncdevv1alpha1.TaskSpec{
						RunConfig: &neosyncdevv1alpha1.RunConfig{
							Benthos: &neosyncdevv1alpha1.BenthosRunConfig{
								ConfigFrom: &neosyncdevv1alpha1.ConfigSource{
									SecretKeyRef: &neosyncdevv1alpha1.ConfigSelector{
										Name: TaskName,
										Key:  "benthos.yaml",
									},
								},
							},
						},
					},
				}
				err = k8sClient.Create(ctx, task)
				Expect(err).To(Not(HaveOccurred()))
			}

			By("Checking if the custom resource was successfully created")
			Eventually(func() error {
				found := &neosyncdevv1alpha1.Task{}
				return k8sClient.Get(ctx, typedNamespaceName, found)
			}, time.Minute, time.Second).Should(Succeed())

			By("Reconciling the custom resource created")
			reconciler := &TaskReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err = reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typedNamespaceName,
			})
			Expect(err).To(Not(HaveOccurred()))
		})
	})
})
