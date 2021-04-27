# TCP服务器---Paguma

## 一、基础服务模块（server）
**方法：**

- 启动服务器：基本的服务器开发
    1. 创建addr
    2. 创建listener
    3. 处理客户端的基本业务
    
- 停止服务：做一些资源的回收和状态的回执
- 运行服务器：调用Start()方法，调用之后阻塞处理，在之间可以做今后的一个扩展
- 初始化Server

**属性：**

- name名称
- 监听的IP
- 监听的端口



## 二、链接模块（connection）

**方法：**

- 启动链接Start()
- 停止链接Stop()
- 获取当前链接的conn对象（套接字）
- 得到链接ID
- 得到客户端链接地址和端口
- 发送数据的方法Send()

**属性：**

- socket TCP套接字
- 链接的ID
- 当前链接的状态（是否已关闭）
- 与当前链接所绑定的处理业务方法
- 等待链接被动退出的channel



## 三、基础路由模块（router）

> 这里之所以称之为“基础路由模块”，是因为在这里目前还只实现了支持单路由模式，后续会添加多路由模式的支持

### 1）Request请求封装

> 将链接和数据绑定在一起

**属性：**

- 链接Connection
- 请求数据

**方法：**

- 得到当前链接
- 得到当前数据
- 新建一个Request请求

### 2）Router模块

抽象层定义一个**抽象IRouter**，有以下方法：

- 处理业务之前的方法
- 处理业务的主方法
- 处理业务之后的方法

然后通过一个**具体的BaseRouter**，来实现上述的三个方法，后续可以有新的router去继承BaseRouter，然后对上述方法进行重写。

> 注意：这里用到了**模板方法设计模式**

### 3）paguma继承router模块

该模块有如下功能：

- IServer增添路由添加功能
- Server增添路由成员
- Connection类绑定一个Router成员
- 在Connection调用已经注册的Router处理业务



## 四、全局配置模块

> 这部分目前是使用json进行一个全局配置，考虑到现在后端系统越来越多使用yaml格式文件，后续升级版本可以考虑支持

实现思路：创建一个全局配置模块utils/globalobj.go，然后通过一个init方法读取到用户配置好的application.json到global对象中去。将paguma框架中的硬编码部分替换为配置文件里面的参数

## 五、消息模块（Message）
**属性：**
- 消息的ID
- 消息长度
- 消息的内容

**方法：**
- setter/getter

### 解决粘包问题的封包拆包模块（DataPack)
**针对Message进行TLV格式的封装**
- 写Message的长度
- 写Message的ID
- 写Message的内容

**针对Message进行TLV格式的拆包**
- 先读取固定长度的head，得到消息内容长度和消息类型
- 再根据消息内容的长度，再次进行一次读写，从conn中读取消息的内容

### 将消息封装机制集成到paguma框架中
- 将Message添加到Request属性中
- 修改链接读取数据的机制，将之前的单纯的读取Bytes改成拆包的读取（按照TLV格式读取）
- 给链接提供一个发包机制：将发送的消息进行打包，再发送

## 六、多路由模式
#### 消息管理模块（支持多路由业务api调度管理）
**属性：**
- 集合：消息ID和对应的router关系的表（map）
**方法：**
  - 根据msgID来索引调度路由方法
  - 添加路由方法到map集合中
  
#### 将消息管理模块集成到框架中
- 将server模块中的Router属性替换成MsgHandler属性
- 修改server模块中的AddRouter方法 
- 将Connection模块中的Router属性替换成MsgHandler属性
- Connection之前调度Router的业务，替换成MsgHandler调度，修改StartReader方法

## 七、读写协程分离
读写分离可以使得读业务和写业务模块更加清晰，并且便于在写模块里面添加更多的扩展业务
- 添加一个Reader和Writer之间通信的channel
- 添加一个Write goroutine
- Reader之前是直接发送给客户端，现在改为发送给channel
- 启动Reader和Writer一同工作

## 八、消息队列和多任务机制
- 创建一个消息队列
    - 属性：消息队列、worker工作池的数量
- 创建多任务的工作池并且启动
    - 根据workerPoolSize的数量去创建Worker
    - 每个Worker都应该用一个goroutine去承载，阻塞的等待与当前worker所对应的channel的消息，一旦有消息到来，worker应该处理当前消息对应的业务，即调用DoMsgHandle方法来处理业务
- 将之前的发送消息，改为发送给消息度列和工作池来处理
    - 定义一个方法，将消息发送给消息队列工作池。为了保证每个worker所受到的request任务是均衡的（这里需要用到负载均衡算法，这里暂时采用平均轮询的方式，但是注意，面试可能会问到，所以**务必搞清楚常见的负载均衡算法**）,只需要将这个request请求发送给taskQueue即可。
    - 将消息发送给对应的channel
- 将消息队列机制集成到paguma框架中
    - 开启并调用消息队列以及worker工作池
    - 将从客户端处理的消息，发送给当前的worker工作池来处理

## 九、链接管理模块
- 创建一个链接管理模块
- 将链接管理模块集成到paguma框架中
  - 将ConnManager加入到server模块中
  - 每次成功与客户端建立链接之后，添加链接到ConnManager 
  - 判断当前链接数量是否超过MaxConn
  - 每次与客户端断开链接后，将连接从ConnManager中删除
- 创建链接之后/销毁链接之前所需要处理的一些业务功能补充，提供给用户能够注册Hook函数
  - 给server添加属性：①该server创建链接之后自动调用的hook函数-OnConnStart()；②该server销毁链接之前自动调用的hook函数OnConnStop()
  - 给server添加方法：①注册OnConnStart()钩子函数的方法；②注册OnConnStop()钩子函数的方法；③调用OnConnStart()钩子函数的方法；④调用OnConnStop()钩子函数的方法
    

## 十、链接属性
给connection模块添加一个可以配置属性的功能

**新增的属性：**
- 链接属性集合：map
- 保护链接属性的互斥锁

**新增的方法：**
- 设置链接属性
- 获取链接属性
- 移除链接属性
  




