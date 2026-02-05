// Copyright 2025 BWI GmbH and Artifact Conduit contributors
// SPDX-License-Identifier: Apache-2.0

package envtest

import (
	"errors"
	"io"
	"time"

	"github.com/ironcore-dev/controller-utils/buildutils"
	utilsenvtest "github.com/ironcore-dev/ironcore/utils/envtest"
	utilapiserver "github.com/ironcore-dev/ironcore/utils/envtest/apiserver"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

type ProcessArgs = utilapiserver.ProcessArgs

type Environment struct {
	cfg       *rest.Config
	env       *envtest.Environment
	ext       *utilsenvtest.EnvironmentExtensions
	k8sClient client.Client
	apiServer *utilapiserver.APIServer
	mainPath  string
	extraArgs ProcessArgs
}

func NewEnvironment(mainPath string, crdDirectoryPaths, apiServiceDirectoryPaths []string) (*Environment, error) {
	env := &envtest.Environment{
		CRDDirectoryPaths: crdDirectoryPaths,
	}
	ext := &utilsenvtest.EnvironmentExtensions{
		APIServiceDirectoryPaths:       apiServiceDirectoryPaths,
		ErrorIfAPIServicePathIsMissing: true,
	}

	return &Environment{
		env:      env,
		ext:      ext,
		mainPath: mainPath,
	}, nil
}

func (e *Environment) SetAPIServerExtraArgs(args ProcessArgs) {
	e.extraArgs = args
}

func (e *Environment) Start(scheme *runtime.Scheme, writer io.Writer) (client.Client, error) {
	cfg, err := utilsenvtest.StartWithExtensions(e.env, e.ext)
	if err != nil {
		return nil, err
	}

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, errors.Join(err, e.Stop())
	}

	apiServer, err := utilapiserver.New(cfg, utilapiserver.Options{
		MainPath:     e.mainPath,
		Args:         e.extraArgs,
		BuildOptions: []buildutils.BuildOption{buildutils.ModModeMod},
		ETCDServers:  []string{e.env.ControlPlane.Etcd.URL.String()},
		Host:         e.ext.APIServiceInstallOptions.LocalServingHost,
		Port:         e.ext.APIServiceInstallOptions.LocalServingPort,
		CertDir:      e.ext.APIServiceInstallOptions.LocalServingCertDir,
		Stdout:       writer,
		Stderr:       writer,
	})
	if err != nil {
		return nil, errors.Join(err, e.Stop())
	}

	if err := apiServer.Start(); err != nil {
		return nil, errors.Join(err, e.Stop())
	}

	e.cfg = cfg
	e.k8sClient = k8sClient
	e.apiServer = apiServer

	return k8sClient, nil
}

func (e *Environment) Stop() error {
	var err error
	if e.apiServer != nil {
		err = e.apiServer.Stop()
	}
	if e.ext != nil {
		err = errors.Join(err, utilsenvtest.StopWithExtensions(e.env, e.ext))
	}

	return err
}

func (e *Environment) WaitUntilReadyWithTimeout(timeout time.Duration) error {
	return utilsenvtest.WaitUntilAPIServicesReadyWithTimeout(timeout, e.ext, e.cfg, e.k8sClient, e.k8sClient.Scheme())
}

func (e *Environment) GetRESTConfig() *rest.Config {
	return e.cfg
}
