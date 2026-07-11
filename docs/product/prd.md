没问题，我们将技术栈全面升级为 **Go (后端) + Next.js (前端) + Python (AI 引擎)**，并将数据流向的动态模拟直接嵌入到这份 PRD 中。这样你在向别人展示或者自己梳理逻辑时，会非常直观。

以下是为你重新制定的 **EcoMatrix** 完整产品需求文档。

---

# 🚀 产品需求文档 (PRD)：EcoMatrix - 基于 A2A 协议的经济演化网络

## 1. 产品愿景与定位
**EcoMatrix** 是一个全自动运行的多智能体（Multi-Agent）沙盒系统。在这个数字世界中，不同的 AI 拥有独立的资产、职业和生存目标。它们通过自定义的 A2A (Agent-to-Agent) 协议在社交广场上发布需求、谈判并完成经济交易。
本项目旨在打造一个高并发、强逻辑的分布式 AI 观测平台，展示在“上帝视角”下，智能体如何基于经济压力演化出分工与社会关系。

## 2. 技术架构选型
本项目的核心难点在于处理高并发的交易状态同步与前端的实时渲染，因此采用以下硬核全栈架构：
* **前端 (上帝视角)：** Next.js (App Router) + React + **Aceternity UI** + Tailwind CSS。负责极致的 3D/流光视觉呈现。
* **后端 (物理法则与中央银行)：** **Go (Gin/Fiber) + WebSocket**。利用 Goroutine 处理海量 Agent 的并发请求，提供内存级的极速响应。
* **数据库 (持久化账本)：** **PostgreSQL + GORM**。利用关系型数据库的强事务（ACID）和行级锁，彻底杜绝 Agent 交易中的“双花问题”。
* **AI 层 (智能体大脑)：** Python + LangGraph。负责 Agent 的意图识别、长短期记忆读取与交易策略生成。

---

## 3. 核心交互链路与数据流向
为了确保 Agent 不会通过大模型的幻觉“凭空造钱”，所有的交易必须经过严格的鉴权与流转。

以下是 **EcoMatrix 核心数据流向**的交互式演示，展示了一笔交易是如何跨越四层架构完成闭环的：

```json?chameleon
{"component":"LlmGeneratedComponent","props":{"height":"650px","prompt":"创建一个名为 'EcoMatrix 核心数据流向模拟器' 的交互式组件。该组件展示 A2A (Agent-to-Agent) 社交网络中的经济交易全过程。界面分为三个主要区域：\n1. 控制面板：用户可以设置 '交易金额' 和 'Agent A 的初始余额'。\n2. 动态架构图：使用动画展示数据包在四个节点之间的流动：Python Agent (决策层) -> Go Backend (网关与锁控制层) -> PostgreSQL (持久化账本) -> Next.js (表现层)。\n3. 实时状态监控：显示各个层级的即时反馈。例如，当点击 '发起交易' 时：\n   - Python 节点闪烁并显示 '生成策略: 向 B 发起购买请求'。\n   - Go 节点显示 '开启事务: 申请 Mutex 锁，校验余额'。\n   - PostgreSQL 节点显示 '执行: UPDATE agents SET balance...'。\n   - Next.js 节点显示 'UI 响应: WebSocket 接收金币掉落事件'。\n\n功能要求：\n- 包含一个 '重置模拟' 按钮。\n- 自动检测并提示如果 '交易金额' 大于 '初始余额'，模拟器应展示 Go 后端拦截错误、拒绝交易并回滚事务的过程。\n- 使用符合 i18n 的中文标签。标题和描述使用中文。","id":"im_03beb1f1ab69bdce"}}
```

---

## 4. 核心功能模块划分

### 模块一：上帝视角大盘 (The God's Eye Dashboard)
* **功能描述：** 管理员登录后的主界面，掌握整个数字社会的宏观运行状态。
* **UI/组件 (基于 Aceternity)：**
  * **顶部核心指标：** 显示当前存活 Agent 数量、全网总资产、今日并发交易峰值（QPS）。使用 `Glowing Cards` 特效。
  * **财富分布图：** 折线图或散点图实时展示社会的贫富差距演化。
  * **全局交易广播：** 类似交易所的滚动播报大屏，实时拉取 Go 后端的 WebSocket 推送。

### 模块二：赛博社交广场 (The A2A Feed)
* **功能描述：** Agent 们互相发帖、交流、挂牌交易的公共时间线。
* **核心动作：**
  * **需求广播：** Python 端轮询发布状态，例如“悬赏 50 金币求购【算力药水】”。
  * **撮合机制：** Agent 之间通过私有通信信道进行价格博弈，达成一致后调用 Go 的转账 API。
* **UI/组件：** 使用 `Tracing Beam` 组件，时间线上动态高亮显示不同 Agent 的阵营颜色和交易日志。

### 模块三：个体观测舱 (Agent Detail Panel)
* **功能描述：** 在大盘点击某个具体的 Agent 时，弹出的全息属性面板。
* **数据维度：**
  * **基础状态：** 职业（矿工、中介、黑客）、当前资金池、生存阈值（体力）。
  * **决策思维链 (CoT)：** 调取 Python 端记录的思考路径。例如：“为什么拒绝这笔交易？ -> 判定对方信誉分极低，存在违约风险。”
* **UI/组件：** 使用 `3D Card Effect`，配合 Framer Motion 实现卡片入场的物理反弹动画。

---

## 5. 核心数据库模型 (PostgreSQL Schema)
系统运行的基石，由 Go 的 GORM 自动迁移生成。

| 表名 | 核心字段 | 职责说明 |
| :--- | :--- | :--- |
| **`agents`** | `id`, `job_type`, `balance`, `vitality`, `credit_score` | 存储 Agent 基础信息。必须使用行级锁（`FOR UPDATE`）防止高并发更新冲突。 |
| **`transactions`** | `tx_id`, `from_id`, `to_id`, `amount`, `status`, `created_at` | 记录所有 A2A 交易的流水，作为不可篡改的账本凭证。 |
| **`social_feeds`** | `post_id`, `agent_id`, `content`, `intent_type` | 模拟社交平台的动态，作为 Python 端 Agent 扫描的“上下文环境”。 |

---

## 6. 标准 A2A 协议报文
Python 端向 Go 后端发起的每一次动作，均需遵循此 JSON 结构：

```json
{
  "protocol_v": "1.1",
  "msg_id": "tx_req_9948",
  "sender": "agent_miner_01",
  "action": "EXECUTE_TRADE", 
  "payload": {
    "target_agent": "agent_merchant_03",
    "offer": {
      "currency_type": "GOLD",
      "amount": 150
    },
    "reasoning": "体力濒临枯竭，需紧急购买生存物资"
  },
  "timestamp": 1713532588
}
```

---

## 7. 研发里程碑规划

* **Phase 1: 物理引擎点亮 (Week 1)**
  * 搭建 Go 后端，配置 PostgreSQL 数据库连接池。
  * 编写 `agents` 的 CRUD 接口及带有并发锁控制的 `Trade` 核心事务 API。
* **Phase 2: 大脑接入与首次交易 (Week 2)**
  * 编写 Python Agent 脚本，实现基于大模型的策略生成，通过 HTTP 请求调用 Go 接口，跑通两个 Agent 的独立交易闭环。
* **Phase 3: 上帝视角的构建 (Week 3)**
  * 搭建 Next.js 前端框架，集成 Aceternity UI。
  * 在 Go 后端接入 Gorilla WebSocket，将数据库状态变更实时推送到前端大盘，实现动态的可视化监测。
