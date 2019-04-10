---
layout: post
title: golang:浅析GO逃逸分析
category: golang
tags: golang
description: 本文将探讨Go中指针，堆栈，堆，逃逸分析和值/指针语义背后的机制和设计
keywords: golang, stack, heap, 逃逸分析, 指针
date: 2019-03-31T13:19:54+08:00
score: 5.0
---

## 原文链接

https://www.ardanlabs.com/blog/2017/05/language-mechanics-on-escape-analysis.html

## 前言

这是四部分系列中的第二篇文章，它们将提供对Go中指针，堆栈，堆，转义分析和值/指针语义背后的机制和设计的理解。 这篇文章的重点是堆和逃逸分析。

四部分系列索引:
1. [GO堆栈和指针](https://ljlu1504.github.io/2019/03/30/stacks-and-pointers-on-go)
2. [GO逃避分析](https://ljlu1504.github.io/2019/03/31/escape-analysis-on-go)
3. [GO内存分析](https://ljlu1504.github.io/2019/04/05/memory-profiling-on-go)
4. [GO数据和语义的设计](https://ljlu1504.github.io/2019/04/06/data-and-semantics-on-go)

## 介绍

在这四部分系列的第一篇文章中，我通过使用一个示例来分析GO指针机制的基础知识，其中一个值在goroutine的堆栈中共享。 我没有告诉你的是当你在堆栈中共享一个值时会发生什么。 要理解这一点，您需要了解值可以存在的另一个内存区域：“堆”。 有了这些知识，您就可以开始学习“逃逸分析”了。

逃逸分析是编译器用于确定程序创建的值的位置的过程。 具体来说，编译器执行静态代码分析以确定是否可以在构造它的函数的堆栈帧上放置一个值，或者该值是否必须“转义”到堆。 在Go中，没有关键字或函数可用于指导编译器做出此决定。 只有通过你编写代码来决定这个决定的惯例。

## 堆
除了堆栈之外，堆是存储器的第二个存储区域，用于存储值。 堆不像堆栈那样自我清理，因此使用此内存的成本更高。 主要是，成本与垃圾收集器（GC）相关，垃圾收集器必须参与以保持该区域清洁。 GC运行时，它将使用25％的可用CPU容量。 此外，它可能会创造微秒级的`stop the world`延迟。 拥有GC的好处是您无需担心管理堆内存，这在历史上一直很复杂且容易出错。

堆上的值构成Go中的内存分配。 这些分配对GC施加了压力，因为需要删除不再由指针引用的堆上的每个值。 需要检查和删除的值越多，GC在每次运行时必须执行的工作量就越多。 因此，调步算法一直在努力平衡堆的大小和运行的速度。

## 共享堆栈
在Go中，不允许`goroutine`有一个指针指向另一个`goroutine`堆栈上的内存。 这是因为当堆栈必须增长或缩小时，`goroutine`的堆栈内存可以用新的内存块替换。 如果`runtime`必须跟踪指向其他`goroutine`堆栈的指针，那么`runtime`需要处理的工作将变得太复杂，而且更新那些堆栈上的指针时`stop the world`延迟将会明显增加。

以下是由于堆栈增长而多次改变堆栈地址空间的堆栈示例。 查看第2行和第6行的输出。您将看到Main Stack Frame内的字符串值的地址发生了两次变化。
```go
// main.go
// Number of elements to grow each stack frame.
// Run with 10 and then with 1024
const size = 1024

// main is the entry point for the application.
func main() {
    s := "HELLO"
    stackCopy(&s, 0, [size]int{})
}

// stackCopy recursively runs increasing the size
// of the stack.
func stackCopy(s *string, c int, a [size]int) {
    println(c, s, *s)

    c++
    if c == 10 {
        return
    }

    stackCopy(s, c, a)
}
```
[play](https://play.golang.org/p/pxn5u4EBSI)

输出：
```
0 0x44dfa8 HELLO
1 0x44dfa8 HELLO
2 0x455fa8 HELLO
3 0x455fa8 HELLO
4 0x455fa8 HELLO
5 0x455fa8 HELLO
6 0x465fa8 HELLO
7 0x465fa8 HELLO
8 0x465fa8 HELLO
9 0x465fa8 HELLO
```
## 逃逸机制
只要在函数的`Stack Frame`范围之外共享一个值，它就会被放置（或分配）在堆上。 逃逸分析算法的工作是找到这些情况并保持程序中的完整性。 完整性在于确保对任何值的访问始终准确，一致和高效。

查看此示例以了解逃逸分析背后的基本机制。
```go
package main

type user struct {
    name  string
    email string
}

func main() {
    u1 := createUserV1()
    u2 := createUserV2()

    println("u1", &u1, "u2", &u2)
}

//go:noinline
func createUserV1() user {
    u := user{
        name:  "Bill",
        email: "bill@ardanlabs.com",
    }

    println("V1", &u)
    return u
}

//go:noinline
func createUserV2() *user {
    u := user{
        name:  "Bill",
        email: "bill@ardanlabs.com",
    }

    println("V2", &u)
    return &u
}
```
[play](https://play.golang.org/p/Y_VZxYteKO)

我使用`go：noinline`指令来阻止编译器直接在`main`中内联这些函数的代码。 内联将擦除函数调用并使此示例复杂化。 我将在下一篇文章中介绍内联的副作用。

在清单1中，您将看到一个具有两个不同函数的程序，这些函数创建一个类型为`user`值并将值返回给调用者。 该函数的版本1在返回时使用值语义。这里的值语义是指该函数创建的临时变量`u`正被复制并传递给调用栈，这意味着调用函数正在接收值本身的副本。
```go
16 func createUserV1() user {
17     u := user{
18         name:  "Bill",
19         email: "bill@ardanlabs.com",
20     }
21
22     println("V1", &u)
23     return u
24 }
```
可以看到在第`17`行到第`20`行执行的用户值的构造。然后在第`23`行，将用户值的副本向上传递给调用栈并返回给调用者。 函数返回后，堆栈看起来像这样。

![](/assets/image/golang/golang_escape_analysis_figure1.png)

可以在图`1`中看到，在调用`createUserV1`之后，两个`Stack Frame`中都存在用户值。 

在函数的第2版中，返回时使用了指针语义。这里的指针语义是指该函数创建的`user`值正在被该函数的调用者共享。 调用函数正在接收该值的地址副本。
```go
27 func createUserV2() *user {
28     u := user{
29         name:  "Bill",
30         email: "bill@ardanlabs.com",
31     }
32
33     println("V2", &u)
34     return &u
35 }
```
您可以看到在第`28`到`31`行使用相同的结构文字来构造用户值，但在第`34`行，返回是不同的。 不是将`user`值的副本传递回调用堆栈，而是传递`user`值的地址副本。 基于此，你可能会认为在调用之后堆栈看起来像这样。

![](/assets/image/golang/golang_escape_analysis_figure2.png)

如果你在图`2`中看到的内容真的发生了，那么您就会遇到完整性问题。 指针指向调用堆栈，进入不再有效的内存。 在`main`的下一个函数调用中，指向的内存将被重新构建并重新初始化。这是逃避分析开始保持完整性的地方。 在这种情况下，编译器将判断出在`createUserV2`的堆栈框架内构造`user`值是不安全的，因此它将在堆上构造值。 这个过程将在第`28`行构造`user`值时发生。

## 代码可读性

正如您在上一篇文章中所了解到的，函数可以直接访问其`Stack Frame`内的内存，但访问其`Stack Frame`外的内存需要间接访问。 这意味着访问转义到堆的值也必须通过指针间接完成。记住`createUserV2`的代码。

代码表面的语法隐藏了其中真正发生的事情。 第`28`行声明的变量`u`表示`user`类型的值。 Go中的构造并没有告诉你值在内存中的位置，所以直到第`34`行的返回语句，你知道该值是否需要转义。 这意味着，即使`u`表示`user`类型的值，也必须通过隐藏的指针访问此`user`值。

在函数调用之后，您可以将堆栈理解成这样。

![](/assets/image/golang/golang_escape_analysis_figure3.png)

函数`createUserV2`的`Stack Frame`上的变量`u`表示`Heap`上的值，而不是`Stack`。 这意味着使用`u`来访问值，需要指针访问，而不是直观从代码看到的直接访问。 您可能会想，为什么不让`u`成为指针，因为访问它所代表的值需要使用指针呢？
```go
27 func createUserV2() *user {
28     u := &user{
29         name:  "Bill",
30         email: "bill@ardanlabs.com",
31     }
32
33     println("V2", u)
34     return u
35 }
```
如果您这样做，那么你会损失代码的可读性。 先不看整个函数，而仅仅关注`return`语句。
```
34     return u
35 }
```
这个`return`语句告诉你什么？ 它所说的就是你的副本被传递给调用者。 但是，当您使用＆运算符时，`return`语句会告诉您什么？
```
34     return &u
35 }
```
由于＆运算符，return 语句现在告诉你，你正在共享调用堆栈，因此转移到堆。 这在可读性方面更强大。

下面是另一个使用指针语义构造值会损害可读性的示例。
```go
01 var u *user
02 err := json.Unmarshal([]byte(r), &u)
03 return u, err
```

您必须与第`02`行的`json.Unmarshal`调用共享指针变量才能使此代码生效。 `json.Unmarshal`调用将创建用户值并将其地址分配给指针变量。[查看完整例子](https://play.golang.org/p/koI8EjpeIx)

这段代码说明了什么：
1. 创建`u`设置为零值的指针。
2. 与`json.Unmarshal`函数共享`u`。
3. 返回调用者`u`的副本。

由`json.Unmarshal`函数创建的`u`正与调用者共享, 但是从代码来看这并不明显。

如果在构造过程中使用值语义时，代码可读性会有什么变化呢？
```
01 var u user
02 err := json.Unmarshal([]byte(r), &u)
03 return &u, err
```
这段代码又说明了什么：
1. 创建`u`设置为零值的值。
2. 与`json.Unmarshal`函数共享`u`。
3. 与调用者共享`u`。

一切都很清楚。 第`02`行将`u`值共享到他的调用函数`json.Unmarshal`，第`03`行将函数自己`Stack Frame`中的`u`值共享给调用者。 此共享将导致`u`值转义。

在构造值时使用值语义，并利用＆运算符的可读性来明确如何共享值。

## 编译器分析报告
要查看编译器所做的决定，可以查看编译器的编译过程。 你需要做的就是在`go build`调用中使用`-gcflags`开关和`-m`选项。
有4个级别的`-m`可以使用，但超过`2`个, 编译器给出的信息就会过多而不好理解。 这里使用2级的`-m`。
```go
16 func createUserV1() user {
17     u := user{
18         name:  "Bill",
19         email: "bill@ardanlabs.com",
20     }
21
22     println("V1", &u)
23     return u
24 }

27 func createUserV2() *user {
28     u := user{
29         name:  "Bill",
30         email: "bill@ardanlabs.com",
31     }
32
33     println("V2", &u)
34     return &u
35 }

$ go build -gcflags "-m -m"
./main.go:16: cannot inline createUserV1: marked go:noinline
./main.go:27: cannot inline createUserV2: marked go:noinline
./main.go:8: cannot inline main: non-leaf function
./main.go:22: createUserV1 &u does not escape
./main.go:34: &u escapes to heap
./main.go:34:   from ~r0 (return) at ./main.go:34
./main.go:31: moved to heap: u
./main.go:33: createUserV2 &u does not escape
./main.go:12: main &u1 does not escape
./main.go:12: main &u2 does not escape
```
您可以看到编译器正在检查是否需要逃逸。 编译器说了什么？
从下面这行，编译器说函数`createUserV1`内部对`println`的函数调用并没有导致`u`值逃逸到堆。 编译器必须做这个检查，因为它与println函数共享。
```
./main.go:22: createUserV1 &u does not escape
```

接下来看一下下面这些行。
```
./main.go:34: &u escapes to heap
./main.go:34:   from ~r0 (return) at ./main.go:34
./main.go:31: moved to heap: u
./main.go:33: createUserV2 &u does not escape
```
这些行说，第`31`行与`u`变量相关联的`user`值，由于第`34`行的返回而逃逸。最后一行说的意思与之前第`22`相同，第`33`行的`println`调用不会导致`user`值转义。
阅读这些报告可能会令人困惑，并且根据所讨论的变量类型是`named`类型还是`literal`类型, 编译器的输出可能会稍微改变。

如之前的讨论，将`u`更改为`literal`类型`*`用户，而不是之前的`named`类型`user`，编译器会给出什么结论呢？。

```go
27 func createUserV2() *user {
28     u := &user{
29         name:  "Bill",
30         email: "bill@ardanlabs.com",
31     }
32
33     println("V2", u)
34     return u
35 }

./main.go:30: &user literal escapes to heap
./main.go:30:   from u (assigned) at ./main.go:28
./main.go:30:   from ~r0 (return) at ./main.go:34
```
现在，编译器称由于第`34`行的返回，`u`变量引用的`user`值（`literal`类型`* user`并在第`28`行分配）正在逃逸。

## 结论
变量的构造并不决定该变量的内存地址如何分配。 只有该值如何被共享才能让编译器决定如何对改变了进行内存分配。 无论何时你在调用堆栈中共享一个值，它都会逃逸。 值得逃避还有其他原因，将在下一篇文章中探讨。

这篇文章试图引导您的是为任何给定类型选择值或指针语义的指南。 每种语义都带来了好处和成本。 值语义将值保留在堆栈上，从而降低了GC的压力。 但是，必须存储，跟踪和维护任何给定值的不同副本。 指针语义将值放在堆上，这会给GC带来压力。 但是，它们很有效，因为只需要存储，跟踪和维护一个值。 关键是正确，一致和平衡地使用每个语义。
