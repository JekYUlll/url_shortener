# Go 短链接生成器

[go: 从0到1实现短链接生成器 | urlshortener | golang | echo | sqlc | redis |](https://www.bilibili.com/video/BV1Unz9YiETV)。  
改为 gin + gorm + mysql。

PostgreSQL中，`BIGSERIAL`是自动递增的 64-bit 整型主键，实际上是：`BIGINT + DEFAULT nextval(...) + 序列`

Docker 在 Windows 上处理路径挂载时要求路径必须是绝对路径，真尼玛绝了。  
`docker-compose down` 会停止并删除所有与当前服务相关的 容器，网络，默认卷（如果使用的是匿名卷），但是不会删除已命名的卷。

---

### 其他

docker:
```bash
docker-compose up -d
```

sqlc:

```bash
brew install sqlc
```

migrate:

```bash
brew install golang-migrate
```

windows上似乎要使用绝对路径？
```bash
migrate create -ext=sql -dir=F:/_code/url_shortener/database/migrate -seq init_schema
```