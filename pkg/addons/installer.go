package addons

import (
	"context"
	"fmt"
	"github.com/gofrs/flock"
	"github.com/launchboxio/operator/api/v1alpha1"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/storage/driver"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Installer struct {
	Settings *cli.EnvSettings
	Client   genericclioptions.RESTClientGetter
}

func NewInstaller(client genericclioptions.RESTClientGetter) *Installer {
	return &Installer{
		Settings: cli.New(),
		Client:   client,
	}
}

func (i *Installer) Exists(a *v1alpha1.HelmRef) (*release.Release, error) {
	actionConfig := new(action.Configuration)
	// TODO: Obviously we need a better way to get this configuration
	if err := actionConfig.Init(i.Client, a.Namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return nil, err
	}

	client := action.NewList(actionConfig)
	client.Deployed = true

	results, err := client.Run()
	if err != nil {
		return nil, err
	}

	for _, result := range results {
		// Check if the release exists, and that it was successful
		if result.Name == a.Name && result.Info.Status == release.StatusDeployed {
			return result, nil
		}
	}

	return nil, nil
}

func (i *Installer) Ensure(a *v1alpha1.HelmRef) (*release.Release, error) {

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(i.Client, a.Namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return nil, err
	}

	client := action.NewUpgrade(actionConfig)
	client.Install = true
	client.Namespace = a.Namespace
	//client.Wait = true

	values := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(a.Values), &values); err != nil {
		return nil, err
	}

	chart := fmt.Sprintf("%s/%s", a.Repo, a.Chart)
	cp, err := client.ChartPathOptions.LocateChart(chart, i.Settings)
	if err != nil {
		return nil, err
	}

	chartReq, err := loader.Load(cp)
	if err != nil {
		return nil, err
	}

	if req := chartReq.Metadata.Dependencies; req != nil {
		if err := action.CheckDependencies(chartReq, req); err != nil {
			return nil, err
		}
	}

	if err = i.EnsureNamespace(a.Namespace); err != nil {
		return nil, err
	}

	if client.Install {
		hClient := action.NewHistory(actionConfig)
		hClient.Max = 1
		if _, err := hClient.Run(a.Name); err == driver.ErrReleaseNotFound {
			iClient := action.NewInstall(actionConfig)
			iClient.Namespace = a.Namespace
			//iClient.Wait = true
			iClient.ReleaseName = a.Name
			return iClient.Run(chartReq, values)
		}
	}

	return client.Run(a.Name, chartReq, values)
}

func (i *Installer) InitRepo(c *repo.Entry) error {
	err := os.MkdirAll(filepath.Dir(i.Settings.RepositoryConfig), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Acquire a file lock for process synchronization
	fileLock := flock.New(strings.Replace(i.Settings.RepositoryConfig, filepath.Ext(i.Settings.RepositoryConfig), ".lock", 1))
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		SafeCloser(fileLock, &err)
	}
	if err != nil {
		return err
	}

	b, err := ioutil.ReadFile(i.Settings.RepositoryConfig)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return err
	}

	r, err := repo.NewChartRepository(c, getter.All(i.Settings))
	if err != nil {
		return err
	}

	if _, err := r.DownloadIndexFile(); err != nil {
		return err
	}

	f.Update(c)

	if err := f.WriteFile(i.Settings.RepositoryConfig, 0644); err != nil {
		return err
	}

	return nil
}

func (i *Installer) EnsureNamespace(namespace string) error {
	rc, err := i.Client.ToRESTConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(rc)
	if err != nil {
		return err
	}
	_, err = clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	// A non-IsNotFound error was returned
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		_, err := clientset.CoreV1().Namespaces().Create(context.TODO(), &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		}, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func SafeCloser(fileLock *flock.Flock, err *error) {
	if fileErr := fileLock.Unlock(); fileErr != nil && *err == nil {
		*err = fileErr
		log.Println(fileErr)
	}
}
