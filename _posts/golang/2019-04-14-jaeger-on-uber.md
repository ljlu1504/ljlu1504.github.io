---
layout: post
title: Jaeger:Uber不断发展的分布式跟踪
category: golang
tags: golang
description: 本文将探讨Go中for range, 值/指针语义背后的机制和设计
keywords: golang, for range, pointer, value
date: 2019-04-07T15:19:54+08:00
score: 5.0
coverage: Distributed_Tracing_Header-768x329.png
---

## 原文链接
https://eng.uber.com/distributed-tracing/

分布式跟踪正迅速成为组织用来监视其复杂的、基于微服务的体系结构的工具中的必备组件。在Uber Engineering，我们的开源分布式跟踪系统Jaeger在2016年得到了大规模的内部采用，集成到数百个微服务中，现在每秒记录数千条跟踪。在我们开始新的一年之际，下面是我们如何走到今天的故事，从研究Zipkin这样的现成解决方案，到我们为什么从pull架构转向push架构，以及分布式跟踪将如何在2017年继续发展。

## From Monolith to Microservices
随着Uber的业务呈指数级增长，我们的软件架构也变得复杂起来。一年多以前，在2015年秋季，我们有大约500个微服务。截至2017年初，我们已经有2000多名员工。这部分是由于业务特性(面向用户的特性，如UberEATS和uberrush)以及内部功能(如欺诈检测、数据挖掘和地图处理)的增加。复杂性增加的另一个原因是从大型单体应用程序转移到分布式微服务体系结构。

正如经常发生的那样，进入微服务生态系统本身也带来了挑战。其中包括微服务对系统的可视性的丧失，以及现在服务之间发生的复杂交互。Uber的工程师们知道，我们的技术对人们的生活有着直接的影响。系统的可靠性是至关重要的，但没有可观测性是不可能实现的。传统的监视工具，如度量和分布式日志记录，仍然有它们的位置，但是它们常常不能提供跨服务的可见性。这就是分布式跟踪蓬勃发展的地方。

## Next, Tracing in TChannel

2015年初，我们开始开发一种面向RPC的网络多路复用和帧协议TChannel。协议的设计目标之一是将 Dapper风格的分布式跟踪作为一等公民构建到协议中。为了实现这个目标，TChannel协议规范将跟踪字段定义为二进制格式的一部分。

Tchannel生成跟踪的原型后端体系结构是一个带有自定义收集器、自定义存储和开源Zipkin UI的push模型。如下图所示：

![](/assets/image/golang/3-4-EngBlog-Distributed-Tracing-at-Uber-768x432.png)

分布式跟踪系统在其他主要技术公司(如谷歌和Twitter)的成功是基于RPC框架的可用性(Stubby和[Finagle](http://twitter.github.io/finagle/))分别是在这些公司中广泛使用的。

类似地，TChannel中的开箱即用跟踪功能是向前迈出的一大步。部署的后端原型立即开始接收来自几十个服务的traces消息。使用TChannel正在构建更多的服务，但是全面的生产部署和广泛采用仍然存在问题。原型后端及其基于Riak/Solr的存储在扩展到Uber的流量时出现了一些问题，并且缺少一些查询功能，无法与Zipkin UI正确地互操作。尽管新服务快速采用TChannel，但Uber仍有大量服务未使用TChannel进行RPC;事实上，大多数负责运行核心业务功能的服务都是在没有TChannel的情况下运行的。这些服务是用四种主要编程语言(Node.js, Python, Go and Java)实现的, 他们使用各种不同的框架进行进程间通信。这种技术领域的异质性使得在Uber部署分布式跟踪比在谷歌和Twitter这样的地方要困难得多。

## Building Jaeger in New York City
Uber纽约工程组织成立于2015年初，有两个主要团队:基础设施可观测性方面，以及一切与产品相关的方面(包括UberEATS和UberRUSH)。由于分布式跟踪是一种生产环境监视的形式，因此它非常适合可观察性。

我们组建了一个由两名工程师和两个目标组成的分布式跟踪团队:将现有原型转化为一个完整的生产系统，使所有Uber微服务都可以使用和采用分布式跟踪。
我们还需要一个项目的代码名。命名事物是计算机科学中两个困难问题之一,因此我们围绕Tracing的themes,detectives,和hunting花了几周的时间进行头脑风暴,直到我们选定了这个名字Jaeger(ˈyā-gər),德国的猎人或狩猎服务员。

NYC团队已经有了运行Cassandra集群的操作经验，这是Zipkin后端直接支持的数据库，所以我们决定放弃基于Riak/Solr的原型。
我们在Go中重新实现了收集器来接受TChannel流量，并将其以与Zipkin兼容的二进制格式存储在Cassandra中。
这允许我们使用Zipkin web和查询服务而不做任何修改，还提供了通过自定义tags搜索跟踪的缺失功能。
我们还在每个收集器中构建了一个动态可配置的乘法因子，将入站流量乘以n次，以便使用生产数据对后端进行压力测试。

早期的Jaeger架构仍然依赖于Zipkin UI和Zipkin存储格式。如下图所示：

![](/assets/image/golang/4-5-EngBlog-Distributed-Tracing-at-Uber-768x432.png)

第二步是使跟踪对所有不使用TChannel for RPC的现有服务可用。
接下来的几个月，我们使用Go、Java、Python和Node.js构建客户端库，以支持对任意服务(包括基于http的服务)的检测。
尽管Zipkin后端相当有名和流行，但它在检测方面缺乏良好的表现，尤其是在Java/Scala生态系统之外。
我们考虑了各种开源工具库，但是它们是由不同的人维护的，不能保证网络上的互操作性，通常使用完全不同的api，并且大多数都需要Scribe或Kafka作为报告spans的传输工具。
我们最终决定编写我们自己的库，这些库将进行互操作性集成测试，支持我们需要的传输，最重要的是，用不同的语言提供一致的检测API。
从一开始，我们就构建了所有的客户端库来支持OpenTracing API。

我们在客户端库的最初版本中构建的另一个新特性是能够轮询tracing后端的抽样策略(sampling strategy)。
当一个服务接收请求,没有tracing的元数据,跟踪仪器通常开始一个新的跟踪请求通过生成一个新的随机跟踪ID。然而,大多数生产跟踪系统,特别是那些需要处理像Uber这样超级的规模数据,并不在存储中分析或记录每一个跟踪。
这样做将创建从服务到跟踪后端非常大的流量，可能比服务处理的实际业务流量大几个数量级。
相反，大多数跟踪系统只对一小部分跟踪进行采样，并且只分析和记录这些采样的跟踪。
抽样决策的精确算法就是我们所说的抽样策略。
抽样策略算法的例子包括:

1. 采样所有。这对于测试很有用，但是在生产中很昂贵!
2. 一种概率方法，其中给定的轨迹以一定的固定概率随机抽样。
3. 一种速率限制方法，其中每个时间单元采样X个跟踪数。例如，可以使用漏桶算法的变体。

大多数现有的zipkin兼容的工具库都支持概率抽样，但是它们期望在初始化时配置抽样速率。
这种方法在大规模使用时会导致几个严重的问题:
1. 给定的服务对于采样率对跟踪后端总体流量的影响知之甚少。例如，即使服务本身具有中等的每秒查询速率(QPS)，它也可能调用另一个具有非常高的扇出因子的下游服务，或者使用大量的检测导致过多的跟踪spans。
2. 在Uber，每天的业务量都表现出很强的季节性;更多的人在高峰时间乘车。固定的采样概率对于非高峰流量可能太低，但是对于高峰流量可能太高。

Jaeger客户机库中的轮询功能就是为了解决这些问题而设计的。
通过将有关适当抽样策略的决策转移到跟踪后端，我们让服务开发人员不必猜测适当的抽样率。
这还允许后端随着流量模式的变化动态调整采样率。
下图显示了从收集器到客户机库的反馈循环。

客户机库的第一个版本仍然使用TChannel通过将跟踪跨进程直接提交给收集器来发送跟踪，因此库的发现和路由依赖于Hyperbahn。
这种依赖关系给采用服务跟踪的工程师带来了不必要的麻烦，无论是在基础设施级别上，还是在必须引入服务的额外库的数量上都是如此，这可能会造成依赖地狱([dependency hell](https://en.wikipedia.org/wiki/Dependency_hell))。

我们通过实现jaeger-agent sidecar流程来解决这个问题，该流程作为基础设施组件部署到所有主机，就像收集检测的代理一样。
所有路由和发现依赖项都封装在jaeger-agent中，我们重新设计了客户端库，以报告跟踪spans到本地UDP端口，并在loopback接口上轮询代理以获得采样策略。
新客户端只需要基本的网络库。
这种体系结构的更改朝着我们使用post-trace采样的愿景迈进了一步:在代理中缓冲内存中的跟踪。

当前的Jaeger架构:用Go实现的后端组件，支持OpenTracing标准的四种语言的客户端库，基于响应的web前端，以及基于Apache Spark的后处理和聚合数据管道。 如下图：

![](/assets/image/golang/5-6-EngBlog-Distributed-Tracing-at-Uber-768x432.png)

## Turnkey Distributed Tracing
Zipkin UI是我们在Jaeger中拥有的最后一个第三方软件。
为了与UI兼容，必须以Zipkin Thrift格式在Cassandra中存储span，这限制了我们的后端和数据模型。
特别是，Zipkin模型不支持OpenTracing标准和我们的客户端库中可用的两个重要特性:键值日志API和以更通用的有向无环图(而不仅仅是跨度树)表示的跟踪。
我们决定冒险一试，更新后端数据模型，并编写一个新的UI。
如下所示，新的数据模型支持键值日志记录和span引用。
它还通过避免进程标签在每个span中重复，优化了进程外发送的数据量:

![](/assets/image/golang/6-OpenTracing-data-model-example-768x432.png)

Jaeger数据模型天生支持键值日志记录和span引用。

我们目前正在完成对后端管道的升级，以使用新的数据模型和新的、更好的优化Cassandra模式。
为了发挥新的数据模型优势，我们用Go实现了一个新的Jaeger-query service和一个全新的使用React构建的web UI。
UI的初始版本主要复制Zipkin UI的现有特性，但是它的架构很容易扩展，可以使用新特性和组件，并作为React组件本身嵌入到其他UI中。
例如，用户可以选择多个不同的视图来可视化跟踪结果，例如跟踪持续时间直方图或服务在跟踪中的累计时间:

![](/assets/image/golang/7-Screen-Shot-Search-Results-768x406.png)

Jaeger UI显示跟踪搜索结果。在右上角，持续时间与时间散点图给出了结果和drill-down的可视化表示。

作为另一个例子，可以根据特定的用例查看单个跟踪。默认呈现为时间序列; 其他视图包括有向无环图或关键路径图:

![](/assets/image/golang/8-Screen-Shot-Trace-View-768x391.png)

Jaeger UI显示了单个跟踪的详细信息。在屏幕的顶部是一个跟踪的小地图，它支持在大型跟踪中更容易地导航。

通过用Jaeger自己的组件替换架构中剩余的Zipkin组件，我们将Jaeger定位为一个交钥匙式的端到端分布式跟踪系统。

我们认为检测库本质上是Jaeger的一部分是至关重要的，通过持续集成测试来确保它们与Jaeger后端之间的兼容性和互操作性。
(这种保证在Zipkin生态系统中是不可用的。)
尤其是，所有受支持的语言(目前是Go、Java、Python和Node.js)之间的互操作性，以及所有受支持的传输协议(目前是HTTP和TChannel)，都在由Uber Engineering RPC团队编写的cross -dock框架的帮助下，作为每次拉请求的一部分进行测试。
您可以在Jaeger -client-go cross - dock存储库中找到Jaeger客户机集成测试的详细信息。
目前，所有Jaeger客户端库都是开源的:
[Go](https://github.com/uber/jaeger-client-go)
[Java](https://github.com/uber/jaeger-client-java)
[Node.js](https://github.com/uber/jaeger-client-node)
[Python](https://github.com/uber/jaeger-client-python)

我们正在将后端和UI代码迁移到Github，并计划很快提供完整的Jaeger源代码。如果您对进度感兴趣，请查看主存储库。我们欢迎大家的贡献，也希望其他人能给Jaeger一个机会。虽然我们对目前的进展感到满意，但Uber的分布式追踪还远远没有结束。

Yuri Shkuro是Uber纽约工程办公室的一名软件工程师，现在很有可能正在努力为Jaeger和Uber的[其他开源项目](http://uber.github.io/)做贡献。

更新于2017年4月15日:[Jaeger](https://github.com/jaegertracing/jaeger)现在是正式的开源软件，带有相应的文档。