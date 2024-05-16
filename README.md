# Override

## 这个仓库什么也不能做，请不要盯着我。

VSCode 配置：

```json
    "github.copilot.advanced": {
        "debug.overrideProxyUrl": "http://localhost:8181",
        "debug.chatOverrideProxyUrl": "http://localhost:8181/v1/chat/completions",
        "authProvider": "github-enterprise"
    },
    "github-enterprise.uri": "https://cocopilot.org",
```

其中 `http://localhost:8181` 是你启动的 `override` 服务地址。

JetBrains 配置：
请在自己的.zshrc/.bashrc文件中配置好环境变量参数，然后重启JetBrains工具.
```shell
   ## code 配置
   export AGENT_DEBUG_OVERRIDE_PROXY_URL="http://127.0.0.1:8181"
   ## chat配置
   export AGENT_DEBUG_OVERRIDE_CAPI_URL="http://127.0.0.1:8181/v1"
```

其中 `http://localhost:8181` 是你启动的 `override` 服务地址。

config.json 配置

```json
{
  "bind": "127.0.0.1:8181",
  "proxy_url": "",
  "timeout": 600,
  "codex_api_base": "https://api-proxy.oaipro.com/v1",
  "codex_api_key": "sk-xxx",
  "codex_api_organization": "",
  "codex_api_project": "",
  "chat_api_base": "https://api-proxy.oaipro.com/v1",
  "chat_api_key": "sk-xxx",
  "chat_api_organization": "",
  "chat_api_project": "",
  "chat_model_default": "gpt-4o",
  "chat_model_map": {}
}
```

`organization` 和 `project` 除非你有，且知道怎么回事再填。

`chat_model_map` 是个模型映射的字典。会将请求的模型映射到你想要的，如果不存在映射，则使用 `chat_model_default` 。

1. 理论上，Chat 部分可以使用 `chat2api` ，而 Codex 代码生成部分则不太适合使用 `chat2api` 。
2. 代码生成部分做过延时生成和客户端 Cancel 处理，很有效节省你的Token。
3. 我目前就试了下 `VSCode` ，至于 `JetBrains` 等IDE尚未适配，如果你有相关经验，请告诉我。
4. 项目基于 `MIT` 协议发布，你可以修改，请保留原作者信息。
5. 有什么问题，请在论坛 https://linux.do 讨论，欢迎PR。

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=linux-do/override&type=Date)](https://star-history.com/#linux-do/override&Date)   

