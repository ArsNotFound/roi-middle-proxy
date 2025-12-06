package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/elazarl/goproxy"
	"golang.org/x/net/http/httpproxy"

	"roi-middle-proxy/internal/config"
	"roi-middle-proxy/internal/filter"
	"roi-middle-proxy/internal/logger/sl"
	"roi-middle-proxy/internal/logger/slgoproxy"
)

const (
	debugLevel = "debug"
	infoLevel  = "info"
	warnLevel  = "warn"
	errorLevel = "error"
)

func main() {
	cfgPath := os.Getenv("ROI_CONFIG_PATH")
	if cfgPath == "" {
		log.Fatal("ROI_CONFIG_PATH env must be set")
	}

	cfg, err := config.ParseConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := setupLogger(cfg.LogLevel)
	if err != nil {
		log.Fatal(err)
	}

	logger.Debug("Debug logging enabled", "config", cfg)

	proxy := goproxy.NewProxyHttpServer()
	proxy.Logger = &slgoproxy.LoggerAdapter{Logger: logger}
	proxy.Verbose = true

	if cfg.UpstreamProxy != nil {
		httpProxy := cfg.UpstreamProxy.HttpProxy

		var httpsProxy string
		if cfg.UpstreamProxy.HttpsProxy != "" {
			httpsProxy = cfg.UpstreamProxy.HttpsProxy
		} else {
			httpsProxy = cfg.UpstreamProxy.HttpProxy
		}

		trProxyConfig := &httpproxy.Config{
			HTTPProxy:  httpProxy,
			HTTPSProxy: httpsProxy,
			NoProxy:    cfg.UpstreamProxy.NoProxy,
			CGI:        false,
		}
		trProxyFunc := trProxyConfig.ProxyFunc()

		proxy.Tr = &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return trProxyFunc(req.URL)
			},
		}

		proxy.ConnectDial = proxy.NewConnectDialToProxy(httpsProxy)
	}

	condition, err := filter.DomainInList(cfg.AllowedHosts)
	if err != nil {
		slog.Error("Failed to create whitelist filter", sl.Err(err))
		os.Exit(1)
	}

	onReq := proxy.OnRequest(goproxy.Not(condition))
	onReq.DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		return req, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusForbidden, "403 Forbidden")
	})

	onReq.HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		return goproxy.RejectConnect, host
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("Starting proxy server", "host", cfg.Proxy.Host, "port", cfg.Proxy.Port)
	server := &http.Server{
		Addr:    cfg.Proxy.Host + ":" + strconv.Itoa(cfg.Proxy.Port),
		Handler: proxy,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return
			}
			logger.Error("Failed to start proxy server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down proxy server")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Failed to shutdown proxy server", "error", err)
		os.Exit(1)
	}

	logger.Info("Proxy server shutdown")
}

func setupLogger(level string) (*slog.Logger, error) {
	var slogLevel slog.Level
	switch level {
	case debugLevel:
		slogLevel = slog.LevelDebug
	case infoLevel:
		slogLevel = slog.LevelInfo
	case warnLevel:
		slogLevel = slog.LevelWarn
	case errorLevel:
		slogLevel = slog.LevelError
	default:
		return nil, fmt.Errorf("invalid log level: %s", level)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel}))
	return logger, nil
}
