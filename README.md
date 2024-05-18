# Override

## 这个仓库什么也不能做，请不要盯着我。

### VSCode 配置：

```json
    "github.copilot.advanced": {
        "debug.overrideProxyUrl": "http://127.0.0.1:8181",
        "debug.chatOverrideProxyUrl": "http://127.0.0.1:8181/v1/chat/completions",
        "authProvider": "github-enterprise"
    },
    "github-enterprise.uri": "https://cocopilot.org",
```

### JetBrains等 配置：

按照 coco dash 页面截图配置后，执行对应系统的脚本后重启IDE：
* `scripts/install.sh` 适用于 `macOS` 和 `Linux`
* `scripts/install-all-users.vbs` 适用于 `Windows`，为电脑上所有用户配置，需要有管理员权限。
* `scripts/install-current-user.vbs` 适用于 `Windows`，为当前用户配置，无需管理员权限。
* `scripts/uninstall` 相关脚本与之对应，为卸载配置。

其中 `http://127.0.0.1:8181` 是你启动的 `override` 服务地址。

### config.json 配置

```json
{
  "bind": "127.0.0.1:8181",
  "proxy_url": "",
  "timeout": 600,
  "codex_api_base": "https://api-proxy.oaipro.com/v1",
  "codex_api_key": "sk-xxx",
  "codex_api_organization": "",
  "codex_api_project": "",
  "codex_max_tokens": 4093,
  "chat_api_base": "https://api-proxy.oaipro.com/v1",
  "chat_api_key": "sk-xxx",
  "chat_api_organization": "",
  "chat_api_project": "",
  "chat_max_tokens": 4096,
  "chat_model_default": "gpt-4o",
  "chat_model_map": {}
}
```

`organization` 和 `project` 除非你有，且知道怎么回事再填。

`chat_model_map` 是个模型映射的字典。会将请求的模型映射到你想要的，如果不存在映射，则使用 `chat_model_default` 。

`codex_max_tokens` 可以设置为你希望的最大Token数，你设置的时候最好知道自己在做什么。

`chat_max_tokens` 可以设置为你希望的最大Token数，你设置的时候最好知道自己在做什么。`gpt-4o` 输出最大为 `4096`

可以通过 `OVERRIDE_` + 大写配置项作为环境变量，可以覆盖 `config.json` 中的值。例如：`OVERRIDE_CODEX_API_KEY=sk-xxxx`

### 其他说明
1. 理论上，Chat 部分可以使用 `chat2api` ，而 Codex 代码生成部分则不太适合使用 `chat2api` 。
2. 代码生成部分做过延时生成和客户端 Cancel 处理，很有效节省你的Token。
3. 我目前就试了下 `VSCode` ，至于 `JetBrains` 等IDE尚未适配，如果你有相关经验，请告诉我。
4. 项目基于 `MIT` 协议发布，你可以修改，请保留原作者信息。
5. 有什么问题，请在论坛 https://linux.do 讨论，欢迎PR。

### Star History

[![Star History Chart](https://api.star-history.com/svg?repos=linux-do/override&type=Date)](https://star-history.com/#linux-do/override&Date)   

