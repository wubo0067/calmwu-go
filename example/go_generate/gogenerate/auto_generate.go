/*
 * @Author: calmwu
 * @Date: 2019-12-02 15:55:57
 * @Last Modified by: calmwu
 * @Last Modified time: 2019-12-02 15:58:09
 */

//
package gogenerate

////go:generate genny -in=../generic/generic_queue.go -out=../gen_int_queue.go gen "Something=int"
//go:generate genny -in=../../github.com/wubo0067/calmwu-go/utils/generic_channel.go -out=../gen_num_channel.go -pkg=main gen "ChannelCustomType=int ChannelCustomName=Num"
//go:generate genny -in=../../github.com/wubo0067/calmwu-go/utils/generic_slicetsort.go -out=../gen_moviesort.go -pkg=main gen "SortedSliceObjType=Movie SortedSliceName=MovieSlice"

// 直接执行go generate 就可以生成代码了，弥补了泛型的缺失
