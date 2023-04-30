# template-matching-for-job
a simple demo of template-matching for my future job

## 项目目录
```
.
├── LICENSE
├── README.md
├── conf
│   └── conf.ini    //配置文件
├── draw                                    
│   ├── main.go  //绘制元素标号功能的实现
│   └── result   //存放绘制完成之后的图像
├── go.mod
├── go.sum
├── images       //存放图片
├── location
│   └── main.go    //确认指定坐标属于第几个元素功能的实现
└── match                                   
    ├── handler
    │   └── templateMatch.go    //模板匹配功能
    ├── main.go
    └── router.go       //gin路由注册

```


## 在开始之前。。。
项目在mac平台使用gocv库实现简易的模板匹配应用

请在设备上先完成opencv与gocv库相关的配置

gocv官方配置文档 https://gocv.io/getting-started/


## 模板匹配
接口文档：http://10.161.155.209:10393/shareDoc?issue=0aa4970f87886d705f48a3120fe058b4&target_id=2369afec-4ed9-4abe-9180-6ac1e0ba368c

执行以下命令(config参数可选，默认为../conf/conf.ini)，启动http服务：
```
cd ./match
go run *.go -config=../conf/conf.ini
```

## 绘制标号
该功能会读取conf包中的配置文件，绘制图像后保存至draw/result目录下
执行以下命令(config参数可选，默认为../conf/conf.ini)运行代码
```
cd ./draw
go run main.go -config=../conf/conf.ini
```
-config参数指定配置文件的位置
该命令默认调用异步接口，如果想要调用同步接口，请修改配置参数
```
go run main.go -config=../conf/conf.ini -method=serial
```
### 结果展示
![Image text](https://user-images.githubusercontent.com/89122882/235345616-5cfc54b6-b9ba-471d-b0be-1e810de4cd09.jpg)

## 判断坐标
执行以下命令运行代码
```
cd ./location
go run main.go -x 500 -y 500
```
参数x和y表示坐标的位置，如果想要调用同步接口，请修改配置参数
```
go run main.go -x 500 -y 500 -method=serial
```

## 温馨提示
如果您不是使用命令行运行代码，例如Goland，请注意修改相关路径

例如将conf.ini文件中的save字段由./result/更改为./draw/result/
