# Go 短链接生成器

[go: 从0到1实现短链接生成器 | urlshortener | golang | echo | sqlc | redis |](https://www.bilibili.com/video/BV1Unz9YiETV)。  

改为 gin + gorm + mysql。

HTTP请求 → API层 → Service层 → Repository层（DAO） → 数据库

PostgreSQL中，`BIGSERIAL`是自动递增的 64-bit 整型主键，实际上是：`BIGINT + DEFAULT nextval(...) + 序列`

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