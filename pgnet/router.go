package pgnet

import "paguma/pgiface"

// BaseRouter 实现router时，先嵌入这个BaseRouter基类，然后根据需要对这个基类的方法进行重写就好了
type BaseRouter struct {}

// 下面的三个方法都为空，是因为有的router不需要PreHandle或者PostHandle，
// 因此所有的router全部继承这个BaseRouter的好处就是： 可以不用实现PreHandle或者PostHandle

// PreHandle 在处理业务之前的钩子方法Hook
func (br *BaseRouter)PreHandle(request pgiface.IRequest) {}

// Handle 在处理业务的主方法Hook
func (br *BaseRouter)Handle(request pgiface.IRequest) {}

// PostHandle 在处理业务之后的钩子方法Hook
func (br *BaseRouter)PostHandle(request pgiface.IRequest) {}
