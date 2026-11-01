# envcheck 0.2.0

检查项目需要的环境变量有没有设好。

```
envcheck [--rc PATH]
```

默认读当前目录的 `.envcheckrc`，一行一个变量名；空行跳过，`#` 开头是注释，重复的名字按首见位置保留。

退出码：0 全部已设，1 有缺失，2 rc 文件不可用。

## 测试

```sh
PYTHONPATH=. python3 -m unittest discover -s tests
```

核对 2026-11-01：9 个用例全通过。
