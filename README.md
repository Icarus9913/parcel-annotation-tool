### parcel升级Annotation工具

当parcel从4.0.3升级到4.0.7后，会给应用pod打上"dce.daocloud.io/parcel.ovs.network.status"这个annotation，
要么使用滚动更新重建pod，要么就给pod patch annotation。
---------------------------------



该工具有三种模式，
* 1、默认模式执行后会输出patch命令，需要执行者手动去给pod ptach annotation
* 2、输出patch执行命令到一个名为anno.sh的shell文件，直接执行shell文件即可
``` ./annotationTool --out-to-shell run ```
* 3、自动模式情况下该工具会帮你自动给pod打上annotation
``` ./annotationTool --really-do-it run  ```

------------------
使用方法：
* 执行``` make all ```后将生成annotationTool二进制文件
* 从parcel-server容器里拷贝一个名为etcd-secrets的证书文件夹，放到与annotationTool二进制文件相同的路径下
