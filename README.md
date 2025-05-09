# Go 短链接生成

原作者：[go: 从0到1实现短链接生成器 | urlshortener | golang | echo | sqlc | redis |](https://www.bilibili.com/video/BV1Unz9YiETV)。  
原项目地址：https://github.com/aeilang/urlshortener

进行重写，改为 **gin** + **gorm** + **mysql**，添加布隆过滤器。

---

### 运行前：

- 修改`config/config.yaml.template`内容，并删除`.template`后缀；  

---

### 其他

docker:
```bash
docker-compose up -d
```