package app

import (
	"context"
	"fmt"
	"log"

	// "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
	"github.com/jekyulll/url_shortener/config"
	"github.com/jekyulll/url_shortener/database"
	"github.com/jekyulll/url_shortener/internal/api"
	"github.com/jekyulll/url_shortener/internal/cache"
	"github.com/jekyulll/url_shortener/internal/repository"
	"github.com/jekyulll/url_shortener/internal/service"
	"github.com/jekyulll/url_shortener/pkg/filter"
	"github.com/jekyulll/url_shortener/pkg/shortcode"
	"gorm.io/gorm"
)

type Application struct {
	r           *gin.Engine
	db          *gorm.DB
	redisClinet *cache.RedisCache
	urlService  *service.URLService
	urlHandler  *api.URLHandler
	cfg         *config.Config
	generator   *shortcode.ShortCodeGeneratorImpl
}

func New() *Application {
	return &Application{}
}

func (a *Application) Init(configPath string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	a.cfg = cfg

	db, err := database.NewDB(cfg.Database)
	if err != nil {
		return err
	}
	a.db = db

	redisClinet, err := cache.NewRedisCache(cfg.Redis)
	if err != nil {
		return err
	}
	a.redisClinet = redisClinet

	a.generator = shortcode.NewShortCodeGeneratorImpl(cfg.ShortCode.Length)

	filter := filter.New(a.cfg.Filter.Capacity, a.cfg.Filter.ErrorRate)

	urlRepo := repository.New(a.db)

	a.urlService = service.New(urlRepo, filter, a.generator,
		cfg.App.DefaultDuration, redisClinet, cfg.App.BaseURL)

	a.urlHandler = api.NewURLHandler(a.urlService)

	// TODO
	// TimeOut未设置
	// Log中间件未设置
	// CORS未设置
	// Recover未设置
	r := gin.Default()
	// r.Use()
	r.POST("/api/url", a.urlHandler.CreateURL)
	r.GET(":code", a.urlHandler.RedirectURL)
	r.GET("/", a.urlHandler.DefaultURL)
	a.r = r
	return nil
}

func (a *Application) Run() {
	go a.start()
	go a.tickCleanUp()

	a.shutDown()
}

func (a *Application) start() {
	if err := a.r.Run(a.cfg.Server.Addr); err != nil {
		log.Println(err)
	}
	// 开机时清理一次过期url
	// TODO 似乎没有生效
	go func() {
		if err := a.urlService.DeleteAllExpired(context.Background()); err != nil {
			log.Println(err)
		}
	}()
	// 用 gracehttp 代替 gin 的 Run, 自动优雅退出
	// ! gracehttp 只能在 Linux 上用

	// if err := gracehttp.Serve(&http.Server{
	// 	Addr:         a.cfg.Server.Addr,
	// 	Handler:      a.e,
	// 	WriteTimeout: a.cfg.Server.WriteTimeout,
	// 	ReadTimeout:  a.cfg.Server.ReadTimeout,
	// }); err != nil {
	// 	fmt.Println(err)
	// }
}

func (a *Application) tickCleanUp() {
	ticker := time.NewTicker(a.cfg.App.CleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		if err := a.urlService.DeleteAllExpired(context.Background()); err != nil {
			log.Println(err)
		}
	}
}

// func (a *Application) tickRebuildFilter() {
// 	ticker := time.NewTicker(24 * time.Hour)
// 	defer ticker.Stop()
// 	for range ticker.C {
// 		if err := a.repo.RebuildBloomFilter(context.Background()); err != nil {
// 			// TODO 层级设计错误
// 		}
// 	}
// }

func (a *Application) shutDown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	defer func() {
		sqlDB, err := a.db.DB()
		if err != nil {
			log.Println(err)
		}
		if err := sqlDB.Close(); err != nil {
			log.Println(err)
		}
	}()

	defer func() {
		if err := a.redisClinet.Close(); err != nil {
			log.Println(err)
		}
	}()

	// TODO
	// gin 的优雅退出

	// 5s时间退出
	_, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()
}
