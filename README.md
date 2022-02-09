### parcel升级Annotation工具

当parcel从4.0.3升级到4.0.7后，会给应用pod打上"dce.daocloud.io/parcel.ovs.network.status"这个annotation，
要么使用滚动更新重建pod，要么就给pod patch annotation。
---------------------------------



该工具有两种模式，
* 1、默认模式执行后会输出patch命令，需要执行者手动去给pod ptach annotation
* 2、自动模式情况下该工具会帮你自动给pod打上annotation
``` ./annotation --really-do-it run  ```

------------------
使用方法：
* 执行``` make all ```后将annotationTool二进制文件放入parcel-server pod中执行
