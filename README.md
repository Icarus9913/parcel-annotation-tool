### parcel升级Annotation工具

当parcel从4.0.3升级到4.0.7后，会给应用pod打上"dce.daocloud.io/parcel.ovs.network.status"这个annotation，
要么使用滚动更新重建pod，要么就给pod patch annotation。
---------------------------------


安装方法：
* 步骤1：x86的linux且有golang环境，直接执行``` make all ```后将生成annotationTool二进制文件。  
        或使用Dockerfile生成一个镜像，将该镜像容器里的annotationTool移出来
* 步骤2：从parcel-server容器里拷贝一个名为``` etcd-secrets ``` 的证书文件夹，放到与annotationTool二进制文件相同的路径下
---------------

使用方法:
* 模式一：默认模式执行后会输出patch命令，需要执行者手动去给pod patch annotation  
``` ./annotationTool run ```


* 模式二输出patch执行命令到一个名为anno.sh的shell文件，直接执行shell文件即可.  //注意--out-to-shell参数需放在run之前.  
``` ./annotationTool --out-to-shell run ```  
``` ./anno.sh```

