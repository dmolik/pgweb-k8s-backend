package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	corev1 "k8s.io/api/core/v1"
	pgv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	DB     *badger.DB
}

//+kubebuilder:rbac:groups=acid.zalan.do,resources=postgresqls,verbs=get;list;watch
//+kubebuilder:rbac:resources=secrets,verbs=get;watch;list
//+kubebuilder:rbac:resources=services,verbs=get;watch;list
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	pg := pgv1.Postgresql{}
	err := r.Get(ctx, req.NamespacedName, &pg)
	if err != nil {
		return ctrl.Result{}, err
	}

	name := pg.Spec.ClusterName
	if name == "" {
		name = pg.ObjectMeta.Name
	}
	log.V(0).Info(fmt.Sprintf("found cluster, name=%s, status=%s", name, pg.Status.PostgresClusterStatus))
	oneMinute, err := time.ParseDuration("1m")
	if err != nil {
		return ctrl.Result{}, err
	}
	if pg.Status.PostgresClusterStatus != "Running" {
		log.V(0).Info(fmt.Sprintf("cluster not running, name=%s, status=%s, requeueing", name, pg.Status.PostgresClusterStatus))
		return ctrl.Result{RequeueAfter: oneMinute}, nil
	}
	user, pass, err := r.getPgCreds(ctx, req)
	if err != nil {
		return ctrl.Result{}, err
	}
	host, err := r.getPgHost(ctx, req)
	if err != nil {
		return ctrl.Result{}, err
	}

	n := strings.Split(req.Name, "-")
	key := req.Namespace + "-" + n[0]
	url := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=require", user, pass, host, "postgres")
	err = r.saveToDB(ctx, req, key, url)
	if err != nil {
		return ctrl.Result{}, err
	}

	clusters, err := r.listClusters(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	clustersFromDB, err := r.getAllFromDB(ctx, req)
	if err != nil {
		return ctrl.Result{}, err
	}
	for _, c := range clustersFromDB {
		if notInList(clusters, c) {
			err = r.deleteFromDB(ctx, req, c)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}


	return ctrl.Result{RequeueAfter: oneMinute}, nil
}

func notInList(list []string, item string) bool {
	for _, l := range list {
		if l == item {
			return false
		}
	}
	return true
}

func (r *ClusterReconciler) listClusters(ctx context.Context) ([]string, error) {
	clusters := pgv1.PostgresqlList{}
	err := r.List(ctx, &clusters)
	if err != nil {
		return []string{}, err
	}
	clusternames := []string{}
	for _, cluster := range clusters.Items {
		clusternames = append(clusternames, cluster.ObjectMeta.Namespace + "-" + strings.Split(cluster.ObjectMeta.Name, "-")[0])
	}
	return clusternames, nil
}

func (r *ClusterReconciler) getPgCreds(ctx context.Context, req ctrl.Request) (string, string, error) {
	secret := corev1.Secret{}
	name := "postgres."+req.Name+".credentials.postgresql.acid.zalan.do"
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: req.Namespace}, &secret)
	if err != nil {
		log := log.FromContext(ctx)
		log.Error(err, "unable to get secret", "secret", name)
		return "", "", err
	}
	user := string(secret.Data["username"])
	pass := string(secret.Data["password"])
	return user, pass, nil
}

func (r *ClusterReconciler) getPgHost(ctx context.Context, req ctrl.Request) (string, error) {
	svc := corev1.Service{}
	err := r.Get(ctx, req.NamespacedName, &svc)
	if err != nil {
		log := log.FromContext(ctx)
		log.Error(err, "unable to get host", "service", req.Name)
		return "", err
	}

	return svc.ObjectMeta.Name +"."+svc.ObjectMeta.Namespace+".svc", nil
}

func (r *ClusterReconciler) getAllFromDB(ctx context.Context, req ctrl.Request) ([]string, error) {
	var keys []string
	err := r.DB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			keys = append(keys, string(k))
		}
		return nil
	})
	if err != nil {
		return []string{}, err
	}
	return keys, nil
}

func (r *ClusterReconciler) saveToDB(ctx context.Context, req ctrl.Request, key, url string) error {
	txn := r.DB.NewTransaction(true)
	if err := txn.Set([]byte(key),[]byte(url)); err != nil {
		txn.Discard()
		return err
	}
	err := txn.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (r *ClusterReconciler) deleteFromDB(ctx context.Context, req ctrl.Request, key string) error {
	err := r.DB.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pgv1.Postgresql{}).
		Complete(r)
}
