---
layout: post
title: golang:GO内存分析
category: golang
tags: golang
description: 本文将探讨Go中指针，堆栈，堆，逃逸分析和值/指针语义背后的机制和设计
keywords: golang, stack, heap, 逃逸分析, 指针
date: 2019-04-05T13:19:54+08:00
score: 5.0
coverage: logo_go_web.png
---

## 原文链接
https://www.ardanlabs.com/blog/2017/06/language-mechanics-on-memory-profiling.html

## 前言

这是四部分系列中的第三篇文章，它们将提供对Go中指针，堆栈，堆，转义分析和值/指针语义背后的机制和设计的理解。 这篇文章的重点是内存性能分析。

四部分系列索引:
1. GO堆栈和指针
2. GO逃避分析
3. GO内存分析
4. GO数据和语义的设计

这个本文相关代码的现场演示：GopherCon新加坡（2017年） - 逃逸分析(https://engineers.sg/video/go-concurrency-live-gophercon-sg-2017--1746)

## 介绍

在上一篇文章中，我通过使用一个共享goroutine堆栈值的示例来教授逃逸分析的基础知识。 我没有告诉你的是其他可能导致值逃逸的场景。 为了帮助您，我将调试一个以令人惊讶的内存方式分配的程序。

## 代码
我想更多地了解io包，所以我给了自己一个快速的项目。 给定一个字节流，编写一个可以找到字符串elvis的函数，并将其替换为字符串Elvis的大写版本。 我们这里谈论的是“国王”，所以他的名字应该总是大写。
以下是我解决方案的源代码:
https://play.golang.org/p/n_SzF4Cer4

以下是对应的的benchmark:
https://play.golang.org/p/TnXrxJVfLV

代码有两个不同的函数来解决这个问题。 这篇文章将重点关注algOne函数，因为它使用的是io包。 你可以使用algTwo函数来进行Memory和CPU性能的对比测试。
这是我们将要使用的输入数据和algOne函数期望的输出。
```
Input:
abcelvisaElvisabcelviseelvisaelvisaabeeeelvise l v i saa bb e l v i saa elvi
selvielviselvielvielviselvi1elvielviselvis

Output:
abcElvisaElvisabcElviseElvisaElvisaabeeeElvise l v i saa bb e l v i saa elvi
selviElviselvielviElviselvi1elviElvisElvis
```

以下是algOne函数的代码：
```go
 80 func algOne(data []byte, find []byte, repl []byte, output *bytes.Buffer) {
 81
 82     // Use a bytes Buffer to provide a stream to process.
 83     input := bytes.NewBuffer(data)
 84
 85     // The number of bytes we are looking for.
 86     size := len(find)
 87
 88     // Declare the buffers we need to process the stream.
 89     buf := make([]byte, size)
 90     end := size - 1
 91
 92     // Read in an initial number of bytes we need to get started.
 93     if n, err := io.ReadFull(input, buf[:end]); err != nil {
 94         output.Write(buf[:n])
 95         return
 96     }
 97
 98     for {
 99
100         // Read in one byte from the input stream.
101         if _, err := io.ReadFull(input, buf[end:]); err != nil {
102
103             // Flush the reset of the bytes we have.
104             output.Write(buf[:end])
105             return
106         }
107
108         // If we have a match, replace the bytes.
109         if bytes.Compare(buf, find) == 0 {
110             output.Write(repl)
111
112             // Read a new initial number of bytes.
113             if n, err := io.ReadFull(input, buf[:end]); err != nil {
114                 output.Write(buf[:n])
115                 return
116             }
117
118             continue
119         }
120
121         // Write the front byte since it has been compared.
122         output.WriteByte(buf[0])
123
124         // Slice that front byte out.
125         copy(buf, buf[1:])
126     }
127 }
```
我想知道的是这个函数的执行情况以及它给堆带来的负荷。 我们将运行一个Benchmark测试来了解这一点。


## Benchmarking
这是我编写的benchmark 函数，它调用algOne函数来执行数据流处理。
```go
15 func BenchmarkAlgorithmOne(b *testing.B) {
16     var output bytes.Buffer
17     in := assembleInputStream()
18     find := []byte("elvis")
19     repl := []byte("Elvis")
20
21     b.ResetTimer()
22
23     for i := 0; i < b.N; i++ {
24         output.Reset()
25         algOne(in, find, repl, &output)
26     }
27 }
```
有了这个benchmark函数，我们可以通过使用-bench，-benchtime和-benchmem选项的go test来运行它。
```go
$ go test -run none -bench AlgorithmOne -benchtime 3s -benchmem
BenchmarkAlgorithmOne-8     2000000          2522 ns/op       117 B/op            2 allocs/op
```
运行benchmark测试后，我们可以看到algOne函数正在为每个操作分配2个值，总共117个字节。 这很好，但我们需要知道函数中的哪些代码行导致了这些分配。 要了解这一点，我们需要为此benchmark生成分析数据。

## Profiling
要生成profile数据，我们将再次运行benchmark 测试，但这次使用-memprofile选项来获取memory profile。
```go
$ go test -run none -bench AlgorithmOne -benchtime 3s -benchmem -memprofile mem.out
BenchmarkAlgorithmOne-8     2000000          2570 ns/op       117 B/op            2 allocs/op
```
Benchmark 测试完成后，测试工具生成了两个新文件。
```go
~/code/go/src/.../memcpu
$ ls -l
total 9248
-rw-r--r--  1 bill  staff      209 May 22 18:11 mem.out       (NEW)
-rwxr-xr-x  1 bill  staff  2847600 May 22 18:10 memcpu.test   (NEW)
-rw-r--r--  1 bill  staff     4761 May 22 18:01 stream.go
-rw-r--r--  1 bill  staff      880 May 22 14:49 stream_test.go
```
源代码位于目录memcpu中，它包含stream.go和stream_test.go。其中 algOne函数位于stream.go中，benchmark函数位于stream_test.go中。 生成的两个新文件为mem.out和memcpu.test。 mem.out文件包含profile 数据，以"目录名+test"命名的memcpu.test文件包含我们在查看profile数据时需要访问的符号文件。
通过mem.out和memcpu.test文件，我们现在可以运行pprof工具来研究profile 数据。
```
$ go tool pprof -alloc_space memcpu.test mem.out
Entering interactive mode (type "help" for commands)
(pprof) _
```
使用Go tool，很容易就能进行内存性能分析。在分析内存性能时，您希望使用-alloc_space选项而不是默认的-inuse_space选项。 这个选项将会告诉你每次分配发生的位置，而不管它在您获取profile文件时分配的内存是否还在使用。
在（pprof）提示符下，我们可以使用list命令检查algOne函数。 此命令将正则表达式作为参数来查找要查看的函数。
```
(pprof) list algOne
Total: 335.03MB
ROUTINE ======================== .../memcpu.algOne in code/go/src/.../memcpu/stream.go
 335.03MB   335.03MB (flat, cum)   100% of Total
        .          .     78:
        .          .     79:// algOne is one way to solve the problem.
        .          .     80:func algOne(data []byte, find []byte, repl []byte, output *bytes.Buffer) {
        .          .     81:
        .          .     82: // Use a bytes Buffer to provide a stream to process.
 318.53MB   318.53MB     83: input := bytes.NewBuffer(data)
        .          .     84:
        .          .     85: // The number of bytes we are looking for.
        .          .     86: size := len(find)
        .          .     87:
        .          .     88: // Declare the buffers we need to process the stream.
  16.50MB    16.50MB     89: buf := make([]byte, size)
        .          .     90: end := size - 1
        .          .     91:
        .          .     92: // Read in an initial number of bytes we need to get started.
        .          .     93: if n, err := io.ReadFull(input, buf[:end]); err != nil || n < end {
        .          .     94:       output.Write(buf[:n])
(pprof) _
```
基于此profile，我们现在知道input和[]buf的底层数组正在堆上分配数据。 由于输入是指针变量，因此profile 实际上是指input指针指向的bytes.Buffer值正在申请内存分配。 因此，让我们首先关注input分配并理解它为什么要申请内存分配。
我们可以假设它正在分配，因为函数调用bytes.NewBuffer正在共享它创建调用堆栈的bytes.Buffer值。 但是，flat列中的值（pprof输出中的第一列）的存在告诉我该值是分配的，因为algOne函数正在以导致它转义的方式共享它。
通过对比查看list命令显示的调用algOne的Benchmark函数的内容，我知道flat列表示函数分配。
```
(pprof) list Benchmark
Total: 335.03MB
ROUTINE ======================== .../memcpu.BenchmarkAlgorithmOne in code/go/src/.../memcpu/stream_test.go
        0   335.03MB (flat, cum)   100% of Total
        .          .     18: find := []byte("elvis")
        .          .     19: repl := []byte("Elvis")
        .          .     20:
        .          .     21: b.ResetTimer()
        .          .     22:
        .   335.03MB     23: for i := 0; i < b.N; i++ {
        .          .     24:       output.Reset()
        .          .     25:       algOne(in, find, repl, &output)
        .          .     26: }
        .          .     27:}
        .          .     28:
(pprof) _
```
由于cum列（第二列）中只有一个值，这告诉我Benchmark函数没有直接分配任何内容。 所有分配都是在该循环内进行的函数调用中进行的。 您可以看到这两个列表中函数调用分配的所有内存数相符。
我们仍然不知道为什么bytes.Buffer值正在分配。 此时go build命令中-gcflags“-m -m”选项的可以派上用场。 Profile只能告诉您哪些值正在逃逸，但go build命令可以告诉您具体原因。

## 编译器分析报告
让我们问一下编译器为什么做出逃逸分析的决定。
```
$ go build -gcflags "-m -m"
```
此命令产生大量输出。 我们只需要搜索输出中包含stream.go：83的内容，因为stream.go是包含此代码的文件的名称，而第83行包含bytes.buffer值的构造。 搜索后我们找到6行相关的分析报告。
```
./stream.go:83: inlining call to bytes.NewBuffer func([]byte) *bytes.Buffer { return &bytes.Buffer literal }

./stream.go:83: &bytes.Buffer literal escapes to heap
./stream.go:83:   from ~r0 (assign-pair) at ./stream.go:83
./stream.go:83:   from input (assigned) at ./stream.go:83
./stream.go:83:   from input (interface-converted) at ./stream.go:93
./stream.go:83:   from input (passed to call[argument escapes]) at ./stream.go:93
```
我们发现第一次出现stream.go：83的编译器分析报告很有意思。
```
./stream.go:83: inlining call to bytes.NewBuffer func([]byte) *bytes.Buffer { return &bytes.Buffer literal }
```
它表明bytes.Buffer值没有被转义，它并没有进行函数调用，编译器决定使用内联方式进行函数调用。
所以这是我写的代码片段：
```go
83     input := bytes.NewBuffer(data)
```
由于编译器选择内联bytes.NewBuffer函数调用，我写的代码转换为：
```go
input := &bytes.Buffer{buf: data}
```
这意味着algOne函数直接构造bytes.Buffer值。 那么现在的问题是，是什么导致从algOne函数frame中逃逸？ 答案在我们在上述编译器报告中找到的其他5行中。
```
./stream.go:83: &bytes.Buffer literal escapes to heap
./stream.go:83:   from ~r0 (assign-pair) at ./stream.go:83
./stream.go:83:   from input (assigned) at ./stream.go:83
./stream.go:83:   from input (interface-converted) at ./stream.go:93
./stream.go:83:   from input (passed to call[argument escapes]) at ./stream.go:93
```
这些行告诉我们的是，第93行的代码导致了逃逸。 变量input被分配给interface值。

## Interfaces
我不记得在代码中对interface值进行了分配。 但是，如果你看第93行，就会发现发生了什么。
```go
 93     if n, err := io.ReadFull(input, buf[:end]); err != nil {
 94         output.Write(buf[:n])
 95         return
 96     }
```
对io.ReadFull的调用导致interface分配。 如果查看io.ReadFull函数的定义，可以看到它是如何通过interface类型接受变量input的。
```go
type Reader interface {
      Read(p []byte) (n int, err error)
}

func ReadFull(r Reader, buf []byte) (n int, err error) {
      return ReadAtLeast(r, buf, len(buf))
}
```
似乎将bytes.Buffer地址传递给调用栈并将其存储在Reader interface 中导致逃逸。 我们知道使用接口有成本：分配和间接。 因此，如果不清楚interface 如何使代码更好，您可能不想使用它。 以下是我遵循的一些指导原则，用于判断是否在代码中使用interface。
使用interface：
1. API的用户需要提供实现细节。
2. API具有内部维护所需的多个实现。
3. 可以更改的API部分已经确定，需要解耦。
不使用interface：
1. 为了使用界面。
2. 概括算法。
3. 当用户可以声明自己的接口时。
现在我们可以问自己，这个算法真的需要io.ReadFull函数吗？ 答案是否定的，因为bytes.Buffer类型有一个我们可以使用的方法集。 对函数拥有的值使用方法可以防止分配。

让我们对代码做些修改，删除io包并直接对变量input使用Read方法。
此代码更改消除了导入io包的需要，因此为了保持所有行号相同，我使用空白标识符对io包导入。 这将允许导入保留在列表中。
```go
 12 import (
 13     "bytes"
 14     "fmt"
 15     _ "io"
 16 )

 80 func algOne(data []byte, find []byte, repl []byte, output *bytes.Buffer) {
 81
 82     // Use a bytes Buffer to provide a stream to process.
 83     input := bytes.NewBuffer(data)
 84
 85     // The number of bytes we are looking for.
 86     size := len(find)
 87
 88     // Declare the buffers we need to process the stream.
 89     buf := make([]byte, size)
 90     end := size - 1
 91
 92     // Read in an initial number of bytes we need to get started.
 93     if n, err := input.Read(buf[:end]); err != nil || n < end {
 94         output.Write(buf[:n])
 95         return
 96     }
 97
 98     for {
 99
100         // Read in one byte from the input stream.
101         if _, err := input.Read(buf[end:]); err != nil {
102
103             // Flush the reset of the bytes we have.
104             output.Write(buf[:end])
105             return
106         }
107
108         // If we have a match, replace the bytes.
109         if bytes.Compare(buf, find) == 0 {
110             output.Write(repl)
111
112             // Read a new initial number of bytes.
113             if n, err := input.Read(buf[:end]); err != nil || n < end {
114                 output.Write(buf[:n])
115                 return
116             }
117
118             continue
119         }
120
121         // Write the front byte since it has been compared.
122         output.WriteByte(buf[0])
123
124         // Slice that front byte out.
125         copy(buf, buf[1:])
126     }
127 }
```
当我们针对此修改后的代码运行benchmark 测试时，我们可以看到bytes.Buffer值的分配已经消失。
```
$ go test -run none -bench AlgorithmOne -benchtime 3s -benchmem -memprofile mem.out
BenchmarkAlgorithmOne-8     2000000          1814 ns/op         5 B/op            1 allocs/op
```
我们还发现性能提升约为29％。 代码从2570 ns/op到1814 ns/op。 通过这个修改，我们现在可以专注于为buf slice分配底层Array。 如果我们再次使用profiler来对我们刚刚生成的新profile数据，我们应该能够识别导致剩余分配的原因。
```
$ go tool pprof -alloc_space memcpu.test mem.out
Entering interactive mode (type "help" for commands)
(pprof) list algOne
Total: 7.50MB
ROUTINE ======================== .../memcpu.BenchmarkAlgorithmOne in code/go/src/.../memcpu/stream_test.go
     11MB       11MB (flat, cum)   100% of Total
        .          .     84:
        .          .     85: // The number of bytes we are looking for.
        .          .     86: size := len(find)
        .          .     87:
        .          .     88: // Declare the buffers we need to process the stream.
     11MB       11MB     89: buf := make([]byte, size)
        .          .     90: end := size - 1
        .          .     91:
        .          .     92: // Read in an initial number of bytes we need to get started.
        .          .     93: if n, err := input.Read(buf[:end]); err != nil || n < end {
        .          .     94:       output.Write(buf[:n])
```
剩下的唯一分配是在第89行，这是针对slice的底层Array。

##Stack Frames
我们想知道为什么buf的底层数组正在分配？ 让我们使用-gcflags“-m -m”选项再次运行go build并搜索stream.go:89。
```
$ go build -gcflags "-m -m"
./stream.go:89: make([]byte, size) escapes to heap
./stream.go:89:   from make([]byte, size) (too large for stack) at ./stream.go:89
```
编译器报告称底层Array“对于堆栈来说太大了”。 这条消息非常具有误导性。 并不是底层Array太大，而是编译器在编译时并不知道底层Array的大小。
只有编译器在编译时知道值的大小时，编译器才能将值放在堆栈上。 这是因为每个函数的每个堆栈帧的大小都是在编译时计算的。 如果编译器不知道值的大小，则将其放在堆上。
为了验证这一点，让我们暂时将slice大小硬编码为5并再次运行benchmark测试。
```go
 89     buf := make([]byte, 5)
```
这次我们运行benchmark 测试时，分配已经消失。
```
 $ go test -run none -bench AlgorithmOne -benchtime 3s -benchmem
BenchmarkAlgorithmOne-8     3000000          1720 ns/op         0 B/op            0 allocs/op
```
如果再次查看编译器报告，您将看不到任何内容正在逃逸。
```
$ go build -gcflags "-m -m"
./stream.go:83: algOne &bytes.Buffer literal does not escape
./stream.go:89: algOne make([]byte, 5) does not escape
```
显然我们不能硬编码slice的大小，因此我们需要承受使用此算法时的1次分配。

##Allocations and Performance
比较我们在每次重构中获得的不同性能提升。
```
Before any optimization
BenchmarkAlgorithmOne-8     2000000          2570 ns/op       117 B/op            2 allocs/op

Removing the bytes.Buffer allocation
BenchmarkAlgorithmOne-8     2000000          1814 ns/op         5 B/op            1 allocs/op

Removing the backing array allocation
BenchmarkAlgorithmOne-8     3000000          1720 ns/op         0 B/op            0 allocs/op
```
通过删除bytes.Buffer，我们获得了约`29％`的性能提升。当所有分配被删除时，性能提高了`~33％`。 内存分配是导致应用程序性能受损的一个因素。

## 结论
Go有一些amazing的工具，可以让您理解编译器在进行逃逸分析时所做出的决策。 根据这些信息，您可以重构代码以尽可能使能在stack上分配的内存时，就不要在heap分配内存。 您不大可能编写零heap分配的程序，但是您希望尽可能减少分配。
话虽如此，永远不要将性能代码作为您的首要任务，因为您不想猜测性能。 编写优化正确性的代码作为您的首要任务。 这意味着首先要关注完整性，可读性和简单性。 当你的程序完成功能后，确定程序是否足够快。 如果速度不够快，请使用语言提供的工具来查找和修复性能问题。
