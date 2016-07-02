package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type OAuthProvider struct {
	config   oauth2.Config
	listener net.Listener

	promisesMu sync.Mutex
	promises   map[string]chan *http.Client
}

func (o *OAuthProvider) deliverCode(state, code string) {
	o.promisesMu.Lock()
	defer o.promisesMu.Unlock()

	fmt.Printf("code %q, state %q", code, state)

	promise, ok := o.promises[state]
	if !ok {
		fmt.Fprintf(os.Stderr, "received code for invalid state %q\n", state)
		return
	}

	token, err := o.config.Exchange(context.Background(), code)
	if err != nil {
		close(promise)
		fmt.Fprintln(os.Stderr, "[error]", err)
		return
	}

	client := o.config.Client(context.Background(), token)

	promise <- client
	close(promise)
}

func (o *OAuthProvider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	code := r.FormValue("code")

	if state == "" {
		http.Error(w, "no state received", http.StatusBadRequest)
		return
	}
	if code == "" {
		http.Error(w, "no code received", http.StatusBadRequest)
		return
	}

	go o.deliverCode(state, code)

	fmt.Fprintln(w, "login token received")
}

func (o *OAuthProvider) Close() error {
	if err := o.listener.Close(); err != nil {
		return err
	}
	return nil
}

func (o *OAuthProvider) RequestClient() (string, chan *http.Client) {
	o.promisesMu.Lock()
	defer o.promisesMu.Unlock()

	randState := fmt.Sprintf("st%d", time.Now().UnixNano())

	replyChan := make(chan *http.Client)
	o.promises[randState] = replyChan

	authURL := o.config.AuthCodeURL(randState)

	return authURL, replyChan
}

func NewOAuthProvider(config oauth2.Config, addr, remoteURL string) (*OAuthProvider, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	config.RedirectURL = remoteURL

	provider := &OAuthProvider{
		config:   config,
		listener: l,
		promises: make(map[string]chan *http.Client),
	}

	go http.Serve(l, provider)

	return provider, nil
}
