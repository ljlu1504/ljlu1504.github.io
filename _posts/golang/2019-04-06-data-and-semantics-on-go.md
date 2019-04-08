---
layout: post
title: golang:GO数据和语义的设计
category: golang
tags: golang
description: 本文将探讨Go中指针，堆栈，堆，逃逸分析和值/指针语义背后的机制和设计
keywords: golang, stack, heap, 逃逸分析, 指针
date: 2019-04-06T13:19:54+08:00
score: 5.0
coverage: logo_go_web.png
---

## 原文链接
https://www.ardanlabs.com/blog/2017/06/design-philosophy-on-data-and-semantics.html

## 前言

这是四部分系列中的第四篇文章，它们将提供对Go中指针，堆栈，堆，转义分析和值/指针语义背后的机制和设计的理解。 这篇文章的重点是内存性能分析。

四部分系列索引:
1. GO堆栈和指针
2. GO逃避分析
3. GO内存分析
4. GO数据和语义的设计

## Design Philosophies
`
“值语义将值保留在堆栈上，这减少了垃圾收集器（GC）的压力。 但是，值语义要求存储，跟踪和维护任何给定值的各种副本。 指针语义将值放在堆上，这会给GC带来压力。 但是，指针语义是有效的，因为只需要存储，跟踪和维护一个值。“ 
- Bill Kennedy
`

如果要在整个软件中保持完整性和可读性，则对于给定类型的数据，一致使用值/指针语义至关重要。 为什么？ 因为，如果要在函数之间传递数据时更改数据的语义，则很难维护代码的清晰一致的心理模型。 随着代码库和团队变得越大，越多的错误，数据竞争和副作用将在代码库中不知不觉蔓延开来。

我想从一系列设计理念开始，这些理念将推动选择一种语义而非另一种语义的指导。

## 设计思路(Mental Models)
`
“让我们想象一个项目最终会有一百万行代码或更多代码。 如今这些项目在美国取得成功的可能性非常低 - 远低于50％。 这是有争议的“。 - Tom Love（Objective C的发明者）
`

Tom还提到一盒复印纸可以容纳10万行代码。 花一点时间让它沉入其中。对于那一盒子的代码，你能让百分之几的代码保持相同的设计思路？

我相信要求一个开发人员维护10%（~10k行代码）的代码已经要求相当多了。 但是，如果我们继续为每个开发人员假设~10k行代码，那么需要一个由100名开发人员组成的团队来维护代码库，才可以实现维护达到一百万行代码的代码库。 这意味着需要协调，分组，跟踪和持续沟通的100人。 现在看看你目前的1到10个开发团队。 你用这么小的规模做得怎么样？ 每个开发人员10k行代码，代码库的大小是否与团队规模一致？

## Debugging
`
“最困难的错误是你的设计思路错误，所以你根本看不出问题” - Brian Kernighan
`

我不相信使用调试器有帮助，除非你并不了解代码的设计思路，现在正在浪费努力试图理解问题出在哪。 调试器在被滥用时是可怕的，并且当你对任何可观察的bug的第一反应就是使用调试器时，你知道吗你正在滥用调试器。

如果您在生产中遇到问题，您会什么办？ 没错，查看日志。 如果日志在开发期间不适合您，那么当生产错误发生时，它们肯定不适合您。 日志需要能体现代码库的设计思路，以便您可以读取代码以查找错误。

## Readability
`
“C语言我所见过的在性能和易用性之间具备最佳平衡的语言。 你可以通过相当直接的编程来做你想做的几乎任何事情。你希望机器可以做什么，通过C语言你能非常好实现你的的设计思路。 你可以很好地预测它的运行速度，你明白发生了什么......“ -  Brian Kernighan
`

我相信Brian的这句话同样适用于Go。 保持这种“设计思路”就是一切。 它强调完整性，可读性和简单性。 随着时间的推移，这些是编写良好的，可维护的软件的基石。 编写维护对给定类型数据使用一致的值/指针语义的代码，是实现此目的的重要方法。

## Data Oriented Design
`
“如果您不了解数据，则无法理解问题。 这是因为所有问题都是独特的，并且特定于您正在使用的数据。 当数据发生变化时，您的问题就会发生变化。 当你的问题发生变化时，算法（数据转换）需要随之改变。“ - 比尔肯尼迪
`

仔细想想，您处理的每个问题都是数据转换问题。 您编写的每个函数和您运行的每个程序都会获取一些输入数据并生成一些输出数据。 从这个角度来看，您的软件设计模型就是对这些数据转换的理解（即它们在代码库中的组织和应用方式）。 “少即是多”的态度提供了从更少的层次，陈述，抽象化，更少复杂性和更少工作量方面解决问题的至关重要思路。 这使您和您的团队（面对问题）更容易，而且它也使硬件更容易执行这些数据转换。

##Type (Is Life)
`
"完整性意味着每个分配，每次内存读取和每次内存写入都是准确，一致和高效的。 类型系统对于确保我们具有这种微观的完整性至关重要。" - William Kennedy
`

如果数据驱动您所做的一切，那么表示数据的类型至关重要。 在我的世界“Type Is Life”中，因为类型为编译器提供了确保数据完整性的能力。 类型还驱动并指示语义规则代码必须尊重它所操作的数据。 正确使用值/指针语义, 就需要从正确使用类型开始。

##Data (With Capability)
`
“当方法为数据提供了特定的或合理的能力时，方法就是有效的.” - William Kennedy
`

在他们必须决定方法的接收器类型之前，值/指针语义的概念不会影响Go开发人员。 这是我看到的一个问题：我应该使用值接收器还是指针接收器？ 一旦我听到这个问题，我就知道开发人员没有很好地掌握这些语义。

方法的目的是为数据提供功能。 考虑一下。 一段数据可以具备某些功能。 我一直希望把重点放在数据上，因为它是驱动程序功能的数据。 数据驱动您编写的算法，您实施的封装以及您可以实现的性能。

##Polymorphism
`
“多态性意味着你编写某个程序，它的行为会有所不同，具体取决于它所运行的数据。” -  Tom Kurtz（BASIC的发明者）
`

我喜欢Tom在上面的引文中所说的话。 函数的行为可能会有所不同，具体取决于它所操作的数据。 数据的行为是将函数本身与函数可以接受和使用的具体数据类型分离开来。 这是一个数据具有功能的一个核心原因。 这正是构建和设计能够适应变化的系统与架构的基石。

##Prototype First Approach
`
“除非开发人员对软件的用途非常了解，否则软件很可能会出现问题。 如果开发人员不了解/理解应用程序，那么获得尽可能多的用户输入和体验至关重要。“ -  Brian Kernighan
`

我希望您始终首先关注理解数据转换所需的具体数据和算法，以解决问题。 采用这种原型第一种方法，并编写可以在生产中部署的具体实现（如果这样做是合理和实用的）。 一旦具体实现工作，一旦您了解了哪些有效且无效，请关注重构，通过提供数据功能将实现与具体数据分离。

##Semantic Guidelines
在声明类型时，您必须决定将哪种语义，值或指针。 接受或返回该类型数据的API必须遵守为该类型选择的语义。 API不允许指示或更改语义，他们必须知道什么语义被用于数据并遵从这一点。 这至少(部分地)有助于如何实现大型代码库的一致性。

以下是基本准则：
1. 在声明类型时，您必须确定正在使用的语义。
2. 函数和方法必须尊重给定类型的语义选择。
3. 避免让方法接收器使用与给定类型相对应的语义不同的语义。
4. 避免使用与对应于给定类型的语义不同的语义接受/返回数据的函数。
5. 避免更改给定类型的语义。
`
这些指南有一些例外，其中最大的是unmarshaling。 Unmarshaling始终需要使用指针语义。 Marshaling和unmarshaling 似乎总是例外。
`

对于给定类型，如何选择一个语义？ 这些指南将帮助您回答这个问题。 以下我们将在某些情况下应用指南的例子：

##Built-In Types

Go的内置类型代表数字，文本和布尔数据。 应使用值语义来处理这些类型。 除非有充分的理由，否则不要使用指针来共享这些类型的值。

作为示例，请查看strings包中的这些函数声明。

```go
func Replace(s, old, new string, n int) string
func LastIndex(s, sep string) int
func ContainsRune(s string, r rune) bool
```

所有这些函数都在API的设计中使用值语义。

##Reference Types

引用类型表示语言中的slice，map，interface，function和channel类型。 这些类型应该使用值语义，因为它们被设计为保持堆栈上并最小化堆压力。 它们允许每个函数拥有自己的值副本，而不是每个函数调用都会导致潜在的分配。 这是可能的，因为这些值包含一个指针，该指针在调用之间共享底层数据结构。

除非有充分的理由，否则不要使用指针来共享这些类型的值。 将调用堆栈中的slice或map值共享到Unmarshal函数可能是一个例外。 作为示例，请查看net包中声明的这两种类型。
```go
type IP []byte
type IPMask []byte
```
IP和IPMask类型都基于byte的slice。 这意味着它们都是引用类型，它们应遵循值语义规则。 这是一个名为Mask的方法，它是为接受IPMask值的IP类型声明的。
```go
func (ip IP) Mask(mask IPMask) IP {
    if len(mask) == IPv6len && len(ip) == IPv4len && allFF(mask[:12]) {
        mask = mask[12:]
    }
    if len(mask) == IPv4len && len(ip) == IPv6len && bytesEqual(ip[:12], v4InV6Prefix) {
        ip = ip[12:]
    }
    n := len(ip)
    if n != len(mask) {
        return nil
    }
    out := make(IP, n)
    for i := 0; i < n; i++ {
        out[i] = ip[i] & mask[i]
    }
    return out
}
```
请注意，此方法是一种mutation操作，并使用值语义API样式。 它使用IP值作为接收器，并根据传入的IPMask值创建新的IP值并将其副本返回给调用者。 该方法遵从您为引用类型使用值语义的事实。

这一点，与内置函数append是相同的。

```go
var data []string
data = append(data, "string")
```
函数append使用值语义进行此mutation操作。 您将slice值传递给append，并在mutation后返回新的slice值。

函数unmarshaling总是例外，他需要指针语义。
```go
func (ip *IP) UnmarshalText(text []byte) error {
    if len(text) == 0 {
        *ip = nil
        return nil
    }
    s := string(text)
    x := ParseIP(s)
    if x == nil {
        return &ParseError{Type: "IP address", Text: s}
    }
    *ip = x
    return nil
  }
```

UnmarshalText方法正在实现encoding.TextUnmarshaler接口。 如果没有使用指针语义，它将无法工作。 但这没关系，因为共享一个值通常是安全的。 在函数unmarshaling总是例外，他需要指针语义。之外，如果指针语义用于引用类型，则应该引发标志。

##User Defined Types

这是您需要做出决定的地方。 您必须在声明类型时决定使用何种语义。
如果我要求您为time包编写API并且我给了您这种类型，该怎么办？
```go
type Time struct {
    sec  int64
    nsec int32
    loc  *Location
}
```
你会用什么语义？
在time包中查看此类型的实现以及工厂函数Now。
```go
func Now() Time {
    sec, nsec := now()
    return Time{sec + unixToInternal, nsec, Local}
  }
```
工厂函数是类型最重要的函数之一，因为它告诉您正在选择什么语义。 Now函数清楚地表明值语义正在发挥作用。 此函数创建Time类型的值，并将此值的副本返回给调用者。 共享时间值不是必需的，它们不需要最终在堆上。

另外，查看Add方法，这是一个mutation操作。
```go
func (t Time) Add(d Duration) Time {
    t.sec += int64(d / 1e9)
    nsec := t.nsec + int32(d%1e9)
    if nsec >= 1e9 {
        t.sec++
        nsec -= 1e9
    } else if nsec < 0 {
        t.sec--
        nsec += 1e9
    }
    t.nsec = nsec
    return t
  }
```
您再次可以看到Add方法遵循为该类型选择的语义。 Add方法使用值接收器对其自己的Time值副本进行操作。 它会改变自己的副本并将Time值的新副本返回给调用。
这是一个接受时间值的函数：
```go
func div(t Time, d Duration) (qmod2 int, r Duration) {
```
再次说明，值语义被用于接受Time类型的值。 时间API唯一使用指针语义的是这些Unmarshal相关函数：
```go
func (t *Time) UnmarshalBinary(data []byte) error {
func (t *Time) GobDecode(data []byte) error {
func (t *Time) UnmarshalJSON(data []byte) error {
func (t *Time) UnmarshalText(data []byte) error {
```
大多数情况下，使用值语义时的能力是有限的。 当数据从一个函数传递到另一个函数时，复制数据是不正确或不合理时，需要将对数据的更改隔离为单个值并进行共享。 这是需要使用指针语义的时候。 如果您不是100％确定复制是正确或合理的，那么使用指针语义。

查看os包中File类型的factory函数。
```go
func Open(name string) (file *File, err error) {
    return OpenFile(name, O_RDONLY, 0)
}
```
Open函数返回一个File类型的指针。 这意味着您应该使用指针语义并始终共享File。 将语义从指针更改为值可能对程序造成破坏性影响。 当函数与您共享一个值时，您应该假定您不允许复制指针指向的值。 如果这样做，结果将是不确定的。

查看更多API，您将看到指针语义的一致使用。
```go
func (f *File) Chdir() error {
    if f == nil {
        return ErrInvalid
    }
    if e := syscall.Fchdir(f.fd); e != nil {
        return &PathError{"chdir", f.name, e}
    }
    return nil
}
```
Chdir方法使用指针语义，即使File值从未变异。 该方法必须遵守该类型的语义约定。

```go
func epipecheck(file *File, e error) {
    if e == syscall.EPIPE {
        if atomic.AddInt32(&file.nepipe, 1) >= 10 {
            sigpipe()
        }
    } else {
        atomic.StoreInt32(&file.nepipe, 0)
    }
}
```
这是一个名为epipecheck的函数，它使用指针语义来接受File值。 再次注意，对于File类型的值，指针语义的一致使用。

## 结论
在代码审查中我一直在寻找值/指针语义的一致使用。 它可以帮助您保持代码的一致性和可预测性。 它还帮助每个人保持清晰一致的代码设计思路。 随着代码库和团队变得更大，值/指针语义的一致使用变得更加重要。

Go的惊人之处在于指针和值语义之间的选择超出了接收器和函数参数的声明。 它遍及整个语言，从`for range`到`interfaces`，`function values`和`slices`的机制。 在以后的文章中，我将展示如何在语言的这些不同部分中显示值/指针语义。