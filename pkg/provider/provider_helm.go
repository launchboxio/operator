package provider

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
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/storage/driver"
	"io/ioutil"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type HelmProvider struct {
	HelmRef  v1alpha1.HelmRef
	Settings *cli.EnvSettings
	Client   genericclioptions.RESTClientGetter
}

func (hp *HelmProvider) Install() (Output, error) {
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(hp.Client, hp.HelmRef.Namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return nil, err
	}

	client := action.NewUpgrade(actionConfig)
	client.Install = true
	client.Namespace = hp.HelmRef.Namespace
	//client.Wait = true

	values := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(hp.HelmRef.Values), &values); err != nil {
		return nil, err
	}

	chart := fmt.Sprintf("%s/%s", hp.HelmRef.Repo, hp.HelmRef.Chart)
	cp, err := client.ChartPathOptions.LocateChart(chart, hp.Settings)
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

	//if err = i.EnsureNamespace(hp.HelmRef.Namespace); err != nil {
	//	return nil, err
	//}

	if client.Install {
		hClient := action.NewHistory(actionConfig)
		hClient.Max = 1
		if _, err := hClient.Run(hp.HelmRef.Name); err == driver.ErrReleaseNotFound {
			iClient := action.NewInstall(actionConfig)
			iClient.Namespace = hp.HelmRef.Namespace
			//iClient.Wait = true
			iClient.ReleaseName = hp.HelmRef.Name
			_, err = iClient.Run(chartReq, values)
			return nil, err
		}
	}

	_, err = client.Run(hp.HelmRef.Name, chartReq, values)
	return nil, err
}

func (hp *HelmProvider) InitRepo(c *repo.Entry) error {
	err := os.MkdirAll(filepath.Dir(hp.Settings.RepositoryConfig), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Acquire a file lock for process synchronization
	fileLock := flock.New(strings.Replace(hp.Settings.RepositoryConfig, filepath.Ext(hp.Settings.RepositoryConfig), ".lock", 1))
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		SafeCloser(fileLock, &err)
	}
	if err != nil {
		return err
	}

	b, err := ioutil.ReadFile(hp.Settings.RepositoryConfig)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return err
	}

	r, err := repo.NewChartRepository(c, getter.All(hp.Settings))
	if err != nil {
		return err
	}

	if _, err := r.DownloadIndexFile(); err != nil {
		return err
	}

	f.Update(c)

	if err := f.WriteFile(hp.Settings.RepositoryConfig, 0644); err != nil {
		return err
	}

	return nil
}

func SafeCloser(fileLock *flock.Flock, err *error) {
	if fileErr := fileLock.Unlock(); fileErr != nil && *err == nil {
		*err = fileErr
		log.Println(fileErr)
	}
}
