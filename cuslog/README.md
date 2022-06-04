# cuslog

自定义实现log日志记录

## 功能特性

- 实现DEBUG、INFO、WARN、ERROR、PANIC、FATAL级别的输出
- 实现默认配置，可自定义配置
- 支持输出文件名和行号
- 支持输出本地和文件
- 支持TEXT、JSON输出

### 软件架构

- 定义options类管理logger配置，使用WithOption选项模式实现对默认值的设定，其中Option是func(options2 *options)类型。
  - std.err为标准输出
  - `TextFormatter`是默认的formatter
- 定义entry类管理logger输出配置
- 使用sync.Pool实现并发安全，并使logger能够复用
- logger类定义输出的方法，支持DEBUG和DEBUGF输出方式

### 执行流程
cuslog.INFO() :调用
->
logger.entry().write() :写入行号、函数、格式等信息
->
e.format() :调用格式化
->
e.logger.opt.formatter.Format(e) :可选输出头部信息，格式化信息
->
e.writer（） :将结果输出到标准输出
->
e.release() :释放资源