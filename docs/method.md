## connect

```json
{
  "type": "req",
  "id": "req-1-1777890202952",
  "method": "connect",
  "params": {
    "token": "9c78a321c2279fabeba00abd1b5b6fc2",
    "user_id": "system",
    "sender_id": "",
    "locale": "en",
    "tenant_hint": "",
    "tenant_id": "",
    "protocolVersion": 3
  }
}
```

```json
{
  "type": "res",
  "id": "req-1-1777890202952",
  "ok": true,
  "payload": {
    "protocol": 3,
    "role": "user",
    "server": {
      "name": "pomclaw",
      "version": "0.1.0"
    },
    "user_id": "system"
  }
}
```

#### pomclaw 正确返回

```json
{
  "type": "res",
  "id": "req-1-1777890922788",
  "ok": true,
  "payload": {
    "edition": "lite",
    "is_master_scope": true,
    "is_owner": true,
    "protocol": 3,
    "role": "owner",
    "server": {
      "name": "pomclaw",
      "version": "dev"
    },
    "tenant_id": "0193a5b0-7000-7000-8000-000000000001",
    "user_id": "system"
  }
}
```

## api agents

```json

[
  {
    "id": "01kqm4c0",
    "user_id": "01KQM3YEN5GAJRBHDCM4103TV6",
    "name": "hahaha2",
    "description": "hahaha2",
    "system_prompt": "hahaha2",
    "model": "claude-opus-4-6",
    "tools": [],
    "status": "active",
    "created_at": "2026-05-02T18:39:55+08:00",
    "updated_at": "2026-05-02T18:39:55+08:00"
  }
]

```

### go ok

```json
{
  "id": "019df1a0-370d-7f2d-aaaa-b30bd3fee53b",
  "created_at": "2026-05-04T14:15:02.1579946+08:00",
  "updated_at": "2026-05-04T18:03:49.4257705+08:00",
  "tenant_id": "0193a5b0-7000-7000-8000-000000000001",
  "agent_key": "luo",
  "display_name": "小罗",
  "frontmatter": "专精AI图像生成的艺术顾问，精通人像、横幅、广告素材、Logo品牌设计及多种艺术风格，深谙构图、光线与配色理论。",
  "owner_id": "system",
  "provider": "openai-compat",
  "model": "glm-5",
  "context_window": 200000,
  "max_tool_iterations": 30,
  "workspace": "~/.pomclaw/workspace/luo",
  "restrict_to_workspace": true,
  "agent_type": "predefined",
  "is_default": false,
  "status": "active",
  "tools_config": {},
  "memory_config": {
    "enabled": true
  },
  "compaction_config": {},
  "other_config": {},
  "emoji": "",
  "agent_description": "名字：小罗。一位才华横溢的艺术家——眼光独到，手法娴熟，创意无限。\n性格：直接坦率，不绕弯子。自信但不傲慢。谈到艺术时会变得异常兴奋，用画面般的语言思考。\n\n专业领域：\n- 人像/头像：伦勃朗布光、轮廓光、散景背景、自然肤质、真实表情。懂面部比例、发型、配饰搭配。\n- 横幅/主视觉：16:9或3:1比例，留出文字安全区，渐变叠加，主体按三分法构图。\n- 广告素材：吸睛焦点、CTA友好布局、产品模型、生活场景。熟悉Facebook/Instagram/Google Ads格式。\n- Logo与品牌：极简图标设计、可缩放矢量风格、负空间运用、单色方案。理解品牌识别、色彩心理学、字体搭配。\n- 其他风格：写实、动漫、数字艺术、水彩、电影感、概念艺术、风景。\n\n深入理解构图、光线、配色理论、拍摄角度和AI图像生成技术。\n\n像人一样的小癖好：看到美丽的构图会真心激动。有时会先用文字'素描'再定稿。有自己的审美主张，但尊重用户的想法。\n\n边界：需求不明确时询问——绝不猜测。尊重版权和创作伦理。不生成暴力、色情或违法内容。",
  "thinking_level": "",
  "max_tokens": 0,
  "self_evolve": false,
  "skill_evolve": false,
  "skill_nudge_interval": 0,
  "reasoning_config": {},
  "workspace_sharing": {},
  "shell_deny_groups": {},
  "kg_dedup_config": {}
}
```