# cuszap
封装zap包，实现日志定制化

## 功能特性
- 支持V level通过整形数据调整级别
- 支持WithValues，返回一个携带指定key-value的logger
- log 包提供 WithContext 和 FromContext 用来将指定的 Logger 添加到某个 Context 中和从某个 Context 中获取 Logger。
- log 包提供了 Log.L() 函数，可以很方便的从 Context 中提取出指定的 key-value 对，作为上下文添加到日志输出中。

### 实现方式
- 设置Options
- cuszap.Init(opt)将opt注册到标准的std logger，zap.RedirectStdLog(l)将输出指向新的logger
- 使用zap.Debug对信息进行显示
- 使用zap.With实现WithValue功能
- 使用zap.Name实现WithName功能
- 使用zap.sync实现Flush功能
- 使用context.WithValue实现WithContext
- 使用context.Value实现FromContext
- 使用zapcore.Level对V(lvl)中的lvl进行强制转换，并使用zap.core().Enabled(lvl)判断当前数字层级lvl是否能够正常显示，实现cuszap.V(lvl)功能
- 使用zap.With功能实现对Context进行绑定，可在网络应用中串联调用