package di

// Option 注册选项
type Option struct {
	Name    string // 服务名称
	Replace bool   // 是否替换已存在的服务
}

// NewOptions 创建选项，支持链式调用
func NewOptions(opts ...Option) *Option {
	opt := &Option{}
	for _, o := range opts {
		if o.Name != "" {
			opt.Name = o.Name
		}
		if o.Replace {
			opt.Replace = true
		}
	}
	return opt
}

// WithName 设置服务名称
func WithName(name string) Option {
	return Option{Name: name}
}

// WithReplace 设置是否替换
func WithReplace(replace bool) Option {
	return Option{Replace: replace}
}
