package nsqlookupd

//
//根据Context的命名，指环境、上下文的意思。通俗来讲，就是保存一些运行环境的信息
//从下面的定义可以看出，Context只是包含了NSQLookupd的指针
//

type Context struct {
	nsqlookupd *NSQLookupd
}
