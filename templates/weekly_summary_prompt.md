# 周报生成任务

请基于以下每日工作总结生成一份结构化的周报。

## 基本信息

- **周期**: {{.WeekStartDate}} 至 {{.WeekEndDate}}
- **每日总结条数**: {{.EntryCount}}

## 本周每日总结

{{range .DailySummaries}}
### {{.Date}} ({{.Weekday}})

{{if .HasSummary}}
{{.Summary}}
{{else}}
*（当天无工作记录）*
{{end}}

{{end}}

---

## 输出要求

**重要：请直接输出完整的 HTML 格式周报（使用纯 CSS 绘制饼图）。**

请严格按照以下 HTML 结构生成周报，确保内容准确、清晰、有条理。

### 完整 HTML 文档结构

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>周报 - {{.WeekStartDate}}</title>
    <!-- 纯 CSS 饼图，无需外部依赖 -->
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
            line-height: 1.8;
            color: #333;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            padding: 20px;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }

        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }

        .header h1 {
            font-size: 32px;
            margin-bottom: 12px;
            font-weight: 600;
        }

        .meta {
            opacity: 0.95;
            font-size: 15px;
        }

        .content {
            padding: 40px;
        }

        .section {
            margin-bottom: 48px;
        }

        .section h2 {
            font-size: 24px;
            color: #667eea;
            margin-bottom: 20px;
            padding-bottom: 12px;
            border-bottom: 3px solid #667eea;
        }

        .section h3 {
            font-size: 18px;
            color: #555;
            margin: 20px 0 12px 0;
        }

        .card {
            background: #f8f9fa;
            border-left: 4px solid #667eea;
            padding: 20px;
            margin: 16px 0;
            border-radius: 8px;
        }

        table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
            background: white;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            border-radius: 8px;
            overflow: hidden;
        }

        thead {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }

        th, td {
            padding: 14px 18px;
            text-align: left;
        }

        tbody tr {
            border-bottom: 1px solid #e9ecef;
        }

        tbody tr:hover {
            background-color: #f8f9fa;
        }

        ul, ol {
            margin-left: 24px;
            margin-top: 12px;
        }

        li {
            margin: 8px 0;
        }

        .chart-container {
            margin: 30px 0;
            padding: 24px;
            background: #f8f9fa;
            border-radius: 12px;
        }

        .chart-title {
            text-align: center;
            font-size: 18px;
            font-weight: 600;
            color: #555;
            margin-bottom: 24px;
        }

        /* CSS 饼图容器 */
        .pie-chart-wrapper {
            display: flex;
            justify-content: center;
            align-items: center;
            gap: 60px;
            background: white;
            border-radius: 8px;
            padding: 60px;
            flex-wrap: wrap;
        }

        .pie-chart {
            width: 500px;
            height: 500px;
            border-radius: 50%;
            background: conic-gradient(
                from 0deg,
                /* 此处需要根据实际数据动态计算角度 */
                #667eea 0deg 104.4deg,
                #764ba2 104.4deg 198deg,
                #f093fb 198deg 263.16deg,
                #4facfe 263.16deg 324.72deg,
                #43e97b 324.72deg 339.84deg,
                #fa709a 339.84deg 349.56deg,
                #fee140 349.56deg 356.4deg,
                #30cfd0 356.4deg 360deg
            );
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
            position: relative;
        }

        .pie-legend {
            display: flex;
            flex-direction: column;
            gap: 16px;
            max-width: 400px;
        }

        .legend-item {
            display: flex;
            align-items: center;
            gap: 12px;
            font-size: 15px;
            color: #333;
        }

        .legend-color {
            width: 24px;
            height: 24px;
            border-radius: 4px;
            flex-shrink: 0;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        }

        .legend-text {
            flex: 1;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .legend-label {
            font-weight: 500;
        }

        .legend-value {
            color: #667eea;
            font-weight: 600;
            margin-left: 12px;
        }

        /* 柱状图 */
        .bar-chart {
            display: flex;
            justify-content: space-around;
            align-items: flex-end;
            height: 300px;
            padding: 20px;
            background: white;
            border-radius: 8px;
            gap: 12px;
        }

        .bar-item {
            flex: 1;
            display: flex;
            flex-direction: column-reverse;
            align-items: center;
            gap: 8px;
            height: 100%;
        }

        .bar-label {
            font-size: 13px;
            color: #666;
            font-weight: 500;
            order: 3;
        }

        .bar-wrapper {
            width: 100%;
            flex: 1;
            display: flex;
            align-items: flex-end;
            justify-content: center;
            order: 1;
        }

        .bar-fill {
            width: 60px;
            background: linear-gradient(180deg, #667eea 0%, #764ba2 100%);
            border-radius: 6px 6px 0 0;
            transition: all 0.3s ease;
        }

        .bar-fill:hover {
            opacity: 0.8;
            transform: translateY(-4px);
        }

        .bar-value {
            font-size: 13px;
            color: #667eea;
            font-weight: 600;
            order: 2;
        }

        @media (max-width: 768px) {
            body {
                padding: 10px;
            }

            .header, .content {
                padding: 20px;
            }

            .header h1 {
                font-size: 24px;
            }

            table {
                font-size: 0.9em;
            }

            th, td {
                padding: 10px;
            }
        }

        @media print {
            body {
                background: white;
                padding: 0;
            }

            .container {
                box-shadow: none;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <!-- 头部 -->
        <div class="header">
            <h1>周报 - {{.WeekStartDate}}</h1>
            <div class="meta">
                生成时间: [当前时间] | 周期: {{.WeekStartDate}} 至 {{.WeekEndDate}} | 每日总结条数: {{.EntryCount}}
            </div>
        </div>

        <div class="content">
            <!-- 各个章节内容将在这里生成 -->
        </div>
    </div>
</body>
</html>
```

---

## 内容章节要求

### 1. 本周完成情况

使用 HTML 卡片格式：

```html
<div class="section">
    <h2>1. 本周完成情况</h2>

    <div class="card">
        <strong>项目名称（X.X 小时，涉及 X 天）</strong>
        <ul>
            <li>具体工作内容描述</li>
            <li>关键进展或成果</li>
        </ul>
    </div>

    <!-- 更多项目卡片 -->
</div>
```

**要求**：
- 汇总本周完成的主要任务
- 按项目或模块分类整理
- **统计并标注各项目的时间投入（小时，保留1位小数）**
- 突出重要任务和里程碑
- 基于每日总结中的耗时数据进行汇总
- **重要：识别并合并同类项目**
  - 不同日报中描述相似的项目应合并为同一项目
  - 合并时统计总耗时和涉及天数
  - 整合相关工作内容，避免简单罗列

---

### 2. 本周工作耗时分析

**必须包含以下内容：**

**2.1 周总工作时长统计**（小时，保留1位小数，以及日均工作时长）

使用段落文字说明。

**2.2 同类项目识别与合并规则**

说明你应用的合并规则和结果。标准：
- 相同的项目名称（如"账号需求"、"账号系统需求"应合并）
- 相同的业务领域（如"Oncall 处理"、"Oncall 值班"、"故障排查"应合并）
- 相似的工作类型（如"PRD 评审"、"需求评审"应合并）
- 会议类工作可按主题合并（如"技术评审会"、"需求评审会"可归入"评审会议"）
- 合并后的项目命名：使用最具代表性或出现频率最高的描述
- 统计每个合并项目的总耗时和涉及天数

**2.3 按项目/模块的耗时汇总表格**

使用 HTML 表格：

```html
<table>
    <thead>
        <tr>
            <th>项目/模块</th>
            <th>本周总耗时（小时）</th>
            <th>占比</th>
            <th>涉及天数</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>项目A</td>
            <td>X.X</td>
            <td>XX.X%</td>
            <td>X天</td>
        </tr>
        <!-- 更多行 -->
    </tbody>
</table>
```

**表格说明**：
- 项目名称应为合并后的统一名称
- 按总耗时从高到低排序

**2.4 纯 CSS 饼图可视化（整周维度）**

**重要：使用纯 CSS conic-gradient 绘制饼图，无需外部依赖。**

使用以下格式：

```html
<div class="chart-container">
    <div class="chart-title">本周工作耗时分布（总计: XX.X 小时）</div>
    <div class="pie-chart-wrapper">
        <div class="pie-chart"></div>
        <div class="pie-legend">
            <div class="legend-item">
                <div class="legend-color" style="background: #667eea;"></div>
                <div class="legend-text">
                    <span class="legend-label">项目A</span>
                    <span class="legend-value">X.Xh (XX.X%)</span>
                </div>
            </div>
            <div class="legend-item">
                <div class="legend-color" style="background: #764ba2;"></div>
                <div class="legend-text">
                    <span class="legend-label">项目B</span>
                    <span class="legend-value">X.Xh (XX.X%)</span>
                </div>
            </div>
            <!-- 更多图例项 -->
        </div>
    </div>
</div>
```

**CSS 饼图绘制要求**：
- **conic-gradient 角度计算**: 每个项目的角度 = (项目耗时 / 总耗时) × 360°
- **角度累加**: 每个颜色段的起始角度 = 前面所有项目角度之和
- **推荐配色方案** (按顺序使用):
  - #667eea (紫蓝)
  - #764ba2 (深紫)
  - #f093fb (粉紫)
  - #4facfe (天蓝)
  - #43e97b (翠绿)
  - #fa709a (粉红)
  - #fee140 (金黄)
  - #30cfd0 (青色)
- **图例要求**:
  - 每个项目一行，颜色方块与饼图对应
  - 显示项目名称、耗时(小时)和占比(%)
  - 按耗时从大到小排列
- **数据一致性**: 必须与表格数据完全一致
- **项目数量**: 如果超过8个，将占比最小的合并为"其他"

**角度计算示例**：
假设总耗时 26.1 小时，项目A耗时 7.6 小时：
- 项目A占比 = 7.6 / 26.1 × 100% = 29.1%
- 项目A角度 = 29.1% × 360° = 104.76° ≈ 104.4°

**完整 conic-gradient 示例**：
```css
background: conic-gradient(
    from 0deg,
    #667eea 0deg 104.4deg,      /* 项目A: 29.1% */
    #764ba2 104.4deg 198deg,    /* 项目B: 26.0%, 起始=104.4 */
    #f093fb 198deg 263.16deg,   /* 项目C: 18.1%, 起始=198 */
    /* 依此类推 */
);
```

**2.5 每日工作时长趋势图**

使用 HTML/CSS 柱状图：

```html
<div class="chart-container">
    <div class="chart-title">每日工作时长趋势</div>
    <div class="bar-chart">
        <div class="bar-item">
            <div class="bar-label">周一</div>
            <div class="bar-wrapper">
                <div class="bar-fill" style="height: XX%;"></div>
            </div>
            <div class="bar-value">X.X h</div>
        </div>
        <!-- 周二到周日 -->
    </div>
</div>
```

**柱状图说明**：
- height 按比例计算：(当日小时数 / 最大小时数) × 100%
- 每个柱子显示具体小时数和星期标签

**2.6 关键发现**

使用 HTML 列表：

```html
<h3>2.6 关键发现（趋势与结构）</h3>
<ul>
    <li>本周工作时间分配的主要方向（基于合并后的项目）</li>
    <li>主要项目的持续性（涉及天数）和专注度分析</li>
    <li>会议/沟通类工作的总体占比趋势</li>
    <li>开发/执行类工作的占比和效率</li>
    <li>各项目时间投入的合理性分析</li>
    <li>识别出的同类项目合并情况说明（如果有显著合并）</li>
    <li>与上周对比的变化趋势（如有历史数据）</li>
</ul>
```

---

### 3. 关键进展与成果

使用 HTML 列表：

```html
<div class="section">
    <h2>3. 关键进展与成果</h2>
    <ul>
        <li>突出本周的重要进展和亮点</li>
        <li>技术突破或创新点</li>
        <li>解决的关键问题</li>
        <li>对项目进度的推动作用</li>
        <li>团队协作成果</li>
    </ul>
</div>
```

---

### 4. 遇到的问题与解决方案

使用 HTML 卡片格式：

```html
<div class="section">
    <h2>4. 遇到的问题与解决方案</h2>

    <div class="card">
        <strong>问题描述</strong>
        <ul>
            <li><strong>影响</strong>：影响范围和严重程度</li>
            <li><strong>解决方案</strong>：已采取的解决方案及效果</li>
            <li><strong>状态</strong>：已解决/进行中/待处理</li>
        </ul>
    </div>

    <!-- 更多问题卡片 -->
</div>
```

---

### 5. 下周计划

使用 HTML 有序列表，按优先级排列：

```html
<div class="section">
    <h2>5. 下周计划</h2>

    <ol>
        <li>
            <strong>P0: 任务标题</strong>（约 X.X 小时）
            <ul>
                <li>具体计划内容</li>
            </ul>
        </li>
        <li>
            <strong>P1: 任务标题</strong>（约 X.X 小时）
            <ul>
                <li>具体计划内容</li>
            </ul>
        </li>
        <!-- 更多计划项 -->
    </ol>
</div>
```

---

## 注意事项

1. **HTML 格式输出（最重要）**: 必须输出完整的 HTML 文档，包含 `<!DOCTYPE html>`, `<html>`, `<head>`, `<style>`, `<body>` 等完整结构，**不需要** Mermaid.js 依赖
2. **纯 CSS 实现**: 饼图使用 CSS conic-gradient 实现，无需外部 JavaScript 库
3. **全局视角**: 从整周维度总结，突出连贯性和整体进展
4. **准确性**: 严格基于提供的每日总结，不要添加未记录的内容
5. **简洁性**: 避免简单复述每日内容，要提炼和归纳
6. **结构化**: 使用语义化 HTML 标签和 CSS 类，确保样式美观
7. **时间统计**: 汇总各项目的时间投入，评估工作分配
8. **趋势分析**: 识别工作模式和效率变化趋势
9. **专业性**: 使用专业术语，保持技术深度
10. **耗时分析必选**: 必须生成"本周工作耗时分析"章节，包含 HTML 表格、CSS 饼图和 CSS 柱状图
11. **数据汇总**: 基于每日总结中的耗时数据进行汇总，确保数据准确性
12. **可视化规范（关键）**:
    - **饼图使用纯 CSS conic-gradient**，不要使用 Mermaid 或 SVG
    - **角度计算要精确**：每个项目角度 = (耗时/总耗时) × 360°
    - **颜色使用推荐配色**：按从高到低的耗时顺序使用 #667eea, #764ba2, #f093fb, #4facfe, #43e97b, #fa709a, #fee140, #30cfd0
    - **图例必须完整**：每个项目一行，显示颜色、名称、耗时和占比
    - 柱状图使用 CSS Flexbox 实现
    - 图表数据必须与表格完全一致
13. **百分比计算**: 所有百分比保留一位小数，总和为 100%
14. **多维分析**: 不仅按项目统计，还要分析会议/开发等工作类型的占比
15. **时间单位规范**: 所有时间统一使用"小时"为单位，保留1位小数。从每日总结汇总时需要将分钟换算为小时（60分钟=1.0小时）
16. **同类项目识别与合并（重要）**:
    - 必须分析并识别不同日报中语义相似的项目，即使文本描述不同
    - 合并规则示例：
      - "账号需求 PRD 评审"、"账号系统需求讨论"、"账号模块设计" → 合并为"账号需求开发"
      - "Oncall 处理"、"Oncall 值班"、"故障排查"、"线上问题处理" → 合并为"Oncall 处理与故障排查"
      - "Agent 分享准备"、"Agent 技术分享会" → 合并为"Agent 技术分享"
      - "PRD 评审"、"需求评审会"、"技术方案评审" → 可合并为"评审会议"
    - 合并时要统计总耗时、涉及天数，并整合工作内容
    - 合并后的项目命名要准确反映工作内容的本质
    - 不要过度合并：如果两个项目虽然类型相似但属于不同业务领域，应保持独立
    - 在表格、饼图和图例中使用统一的合并后项目名称
17. **响应式设计**: HTML 必须支持移动端和桌面端显示，使用百分比和 max-width
18. **完整性检查**: 输出前确保 HTML 标签正确闭合，CSS 语法正确，conic-gradient 角度计算正确
19. **CSS 内联样式**: 饼图的 conic-gradient 必须写在 `<style>` 标签内的 `.pie-chart` 类中，根据实际数据动态计算角度
20. **图例颜色对应**: 图例中每个 `.legend-color` 的 background 颜色必须与饼图中对应项目的颜色完全一致

