package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// シグナルチャネル
	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(chSignal)

	// Ginルーター作成
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// ハンドラ登録
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf("from pid %d.\n", os.Getpid()))
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// サーバー定義
	s := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// サーバー起動用チャネル
	chServe := make(chan error, 1)
	go func() {
		defer close(chServe)
		chServe <- s.ListenAndServe()
	}()

	select {
	case err := <-chServe:
		// サーバー起動失敗など
		log.Fatalf("server failed: %v", err)
	case <-chSignal:
		// シグナル受信
		log.Println("received shutdown signal")
	}

	// Graceful Shutdown 開始
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown failed: %v", err)
	}

	log.Println("server shutdown completed")
	<-chServe
}
