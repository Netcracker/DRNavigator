package service

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	envconfig "github.com/netcracker/drnavigator/site-manager/config"
	"github.com/netcracker/drnavigator/site-manager/pkg/model"
	k8sauth "k8s.io/api/authentication/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var twLog = ctrl.Log.WithName("token-watcher")

const smServiceAccountName = "sm-auth-sa"

// TokenWatcher is interface to validate sm-auth-sa token and watch the token from specified path
type TokenWatcher interface {
	GetToken() string
	ValidateToken(ctx context.Context, token string) (bool, error)
	Start() error
}

// tokenWatcher is implementation of TokenWatcher
type tokenWatcher struct {
	sync.RWMutex

	SmConfig   *model.SMConfig
	KubeClient client.Client
	tokenPath  string

	currentToken string
	watcher      *fsnotify.Watcher
}

// GetToken returns the current token
func (tw *tokenWatcher) GetToken() string {
	if tw.SmConfig.Testing.Enabled {
		return tw.SmConfig.Testing.Token
	}
	tw.RLock()
	defer tw.RUnlock()
	return tw.currentToken
}

// ValidateToken validates, if token is from sm-auth-sa
func (tw *tokenWatcher) ValidateToken(ctx context.Context, token string) (bool, error) {
	if tw.SmConfig.Testing.Enabled {
		return token == tw.SmConfig.Testing.Token, nil
	}

	tokenReview := &k8sauth.TokenReview{
		Spec: k8sauth.TokenReviewSpec{
			Token: token,
		},
	}
	err := tw.KubeClient.Create(ctx, tokenReview)

	if err != nil {
		twLog.Error(err, "There is an error during TokenReview Request")
		return false, err
	}

	if !tokenReview.Status.Authenticated {
		return false, nil
	}

	userName := tokenReview.Status.User.Username
	return fmt.Sprintf("system:serviceaccount:%s:%s", envconfig.EnvConfig.PodNamespace, smServiceAccountName) == userName, nil
}

// ReadToken reads the token from the file
func (tw *tokenWatcher) readToken() error {
	byteData, err := os.ReadFile(tw.tokenPath)
	if err != nil {
		twLog.Error(err, "error reading token from file")
		return err
	}

	tw.Lock()
	tw.currentToken = string(byteData)
	tw.Unlock()
	twLog.V(0).Info("Auth token is updated")
	return nil
}

func (tw *tokenWatcher) Start() error {
	if err := tw.readToken(); err != nil {
		return err
	}

	if err := tw.watcher.Add(tw.tokenPath); err != nil {
		twLog.Error(err, "error adding token file to watch")
		return err
	}

	twLog.V(0).Info("Starting token watcher")

	for {
		select {
		case event := <-tw.watcher.Events:
			if err := tw.handleEvent(event); err != nil {
				_ = tw.watcher.Close()
				return err
			}
		case err := <-tw.watcher.Errors:
			twLog.Error(err, "watch token error")
		}
	}
}

func (tw *tokenWatcher) handleEvent(event fsnotify.Event) error {
	twLog.V(1).Info("watch token event", "event", event)
	if event.Op.Has(fsnotify.Remove) {
		if err := tw.watcher.Add(event.Name); err != nil {
			twLog.Error(err, "error re-watching file")
			return err
		}
		if err := tw.readToken(); err != nil {
			twLog.Error(err, "error re-reading file")
			return err
		}
	} else if event.Op.Has(fsnotify.Create) || event.Op.Has(fsnotify.Write) {
		return tw.readToken()
	}
	return nil
}

// NewTokenWatcher creates the new instance for TokenWatcher
func NewTokenWatcher(smConfig *model.SMConfig, kubeClient client.Client, tokenPath string) (TokenWatcher, error) {
	var err error
	tw := &tokenWatcher{SmConfig: smConfig, KubeClient: kubeClient, tokenPath: tokenPath}

	if tw.watcher, err = fsnotify.NewWatcher(); err != nil {
		return nil, err
	}

	return tw, nil
}
