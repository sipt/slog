# slog
golang demo for output log 

golang log输出的一个demo

参照了beego的log输出方式

主要分成以下几个部分
>用户log传入chan中进行缓存
>打包器:从中取出进行打包，并进行分包判断
>log输出器:打包好的log输出到文件，并根据条件是否分文件，写入完成再去向打包器要Log包裹
