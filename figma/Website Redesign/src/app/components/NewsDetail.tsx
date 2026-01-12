import { ChevronRight, Calendar, User, Eye, Share2, Printer, ArrowLeft } from "lucide-react";

export function NewsDetail() {
  const relatedNews = [
    {
      id: 1,
      title: "2025年12月汽车工业经济运行情况分析",
      date: "2025-12-28",
    },
    {
      id: 2,
      title: "关于开展汽车产业链供应链调研的通知",
      date: "2025-12-25",
    },
    {
      id: 3,
      title: "智能网联汽车标准体系建设指南发布",
      date: "2025-12-20",
    },
  ];

  const hotNews = [
    { title: "新能源汽车下乡活动持续深入", views: 1258 },
    { title: "汽车出口创历史新高", views: 1156 },
    { title: "智能驾驶技术取得新突破", views: 1089 },
    { title: "汽车芯片国产化进程加速", views: 987 },
    { title: "氢能源汽车示范应用启动", views: 856 },
  ];

  return (
    <div className="min-h-screen bg-gray-50">
      {/* 面包屑导航 */}
      <div className="bg-white border-b border-gray-300">
        <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center gap-2 text-sm text-gray-600">
            <a href="#home" className="hover:text-blue-800">首页</a>
            <ChevronRight className="w-4 h-4" />
            <a href="#news" className="hover:text-blue-800">新闻动态</a>
            <ChevronRight className="w-4 h-4" />
            <a href="#association-news" className="hover:text-blue-800">协会动态</a>
            <ChevronRight className="w-4 h-4" />
            <span className="text-gray-900">正文</span>
          </div>
        </div>
      </div>

      <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid lg:grid-cols-4 gap-8">
          {/* 主内容区 */}
          <div className="lg:col-span-3">
            <div className="bg-white border border-gray-300">
              {/* 文章头部 */}
              <div className="border-b-4 border-red-600 p-8 sm:p-12">
                <div className="mb-6">
                  <span className="inline-block px-4 py-1 bg-red-600 text-white text-sm mb-4">
                    协会动态
                  </span>
                </div>
                
                <h1 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-6 leading-tight">
                  2026年1月汽车工业经济运行情况新闻发布会
                </h1>

                {/* 元信息 */}
                <div className="flex flex-wrap items-center gap-6 text-sm text-gray-600 pb-6 border-b border-gray-200">
                  <div className="flex items-center gap-2">
                    <Calendar className="w-4 h-4" />
                    <span>发布时间：2026-01-06 14:30</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <User className="w-4 h-4" />
                    <span>来源：中国汽车工业协会</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Eye className="w-4 h-4" />
                    <span>浏览量：2,458</span>
                  </div>
                </div>

                {/* 操作按钮 */}
                <div className="flex items-center gap-3 mt-6">
                  <button className="flex items-center gap-2 px-4 py-2 border border-gray-300 hover:bg-gray-50 transition-colors text-sm">
                    <Share2 className="w-4 h-4" />
                    分享
                  </button>
                  <button className="flex items-center gap-2 px-4 py-2 border border-gray-300 hover:bg-gray-50 transition-colors text-sm">
                    <Printer className="w-4 h-4" />
                    打印
                  </button>
                </div>
              </div>

              {/* 文章内容 */}
              <div className="p-8 sm:p-12">
                <div className="prose max-w-none">
                  {/* 摘要 */}
                  <div className="bg-blue-50 border-l-4 border-blue-800 p-6 mb-8">
                    <p className="text-gray-800 leading-relaxed font-medium">
                      <strong>摘要：</strong>2026年1月，汽车产销分别完成245.2万辆和242.8万辆，环比分别下降11.9%和14.9%，同比分别增长3.5%和4.2%。其中新能源汽车产销分别完成85.6万辆和84.2万辆，市场占有率达到34.7%。
                    </p>
                  </div>

                  {/* 配图 */}
                  <div className="mb-8">
                    <img
                      src="https://images.unsplash.com/photo-1647427060118-4911c9821b82?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxjYXIlMjBtYW51ZmFjdHVyaW5nJTIwYXNzZW1ibHl8ZW58MXx8fHwxNzY3NzU2MzEzfDA&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral"
                      alt="新闻配图"
                      className="w-full border border-gray-300"
                    />
                    <p className="text-sm text-gray-500 mt-2 text-center">图：汽车生产线</p>
                  </div>

                  {/* 正文内容 */}
                  <div className="space-y-6 text-gray-800 leading-relaxed">
                    <p className="indent-8">
                      1月6日，中国汽车工业协会在北京召开2026年1月汽车工业经济运行情况新闻发布会。会上，中国汽车工业协会副秘书长陈士华发布了1月汽车产销数据。
                    </p>

                    <h2 className="text-2xl font-bold text-gray-900 mt-8 mb-4 pb-3 border-b-2 border-gray-200">
                      一、1月汽车产销总体情况
                    </h2>

                    <p className="indent-8">
                      2026年1月，汽车产销分别完成245.2万辆和242.8万辆，环比分别下降11.9%和14.9%，同比分别增长3.5%和4.2%。
                    </p>

                    <p className="indent-8">
                      1月，乘用车产销分别完成208.6万辆和206.3万辆，环比分别下降13.2%和16.1%，同比分别增长4.2%和5.1%。在乘用车主要品种中，与上月相比，四大类乘用车品种产销均呈下降；与上年同期相比，交叉型乘用车产销呈较快下降，其他品种呈不同程度增长。
                    </p>

                    <h2 className="text-2xl font-bold text-gray-900 mt-8 mb-4 pb-3 border-b-2 border-gray-200">
                      二、新能源汽车产销情况
                    </h2>

                    <p className="indent-8">
                      1月，新能源汽车产销分别完成85.6万辆和84.2万辆，环比分别下降8.5%和11.2%，同比分别增长32.1%和35.8%，市场占有率达到34.7%。
                    </p>

                    <p className="indent-8">
                      在新能源汽车主要品种中，与上月相比，纯电动汽车和插电式混合动力汽车产销均呈下降；与上年同期相比，纯电动汽车和插电式混合动力汽车产销均呈快速增长。
                    </p>

                    <h2 className="text-2xl font-bold text-gray-900 mt-8 mb-4 pb-3 border-b-2 border-gray-200">
                      三、汽车出口情况
                    </h2>

                    <p className="indent-8">
                      1月，汽车企业出口42.5万辆，环比下降15.6%，同比增长18.9%。分车型看，乘用车出口36.8万辆，环比下降16.2%，同比增长22.1%；商用车出口5.7万辆，环比下降11.5%，同比增长3.2%。
                    </p>

                    <div className="bg-gray-50 border border-gray-300 p-6 my-8">
                      <h3 className="text-lg font-bold text-gray-900 mb-4">数据要点</h3>
                      <ul className="space-y-2 text-gray-700">
                        <li className="flex items-start gap-2">
                          <span className="w-1.5 h-1.5 bg-blue-800 rounded-full mt-2 flex-shrink-0"></span>
                          <span>汽车产销分别完成245.2万辆和242.8万辆，同比增长3.5%和4.2%</span>
                        </li>
                        <li className="flex items-start gap-2">
                          <span className="w-1.5 h-1.5 bg-blue-800 rounded-full mt-2 flex-shrink-0"></span>
                          <span>新能源汽车市场占有率达到34.7%，同比增长35.8%</span>
                        </li>
                        <li className="flex items-start gap-2">
                          <span className="w-1.5 h-1.5 bg-blue-800 rounded-full mt-2 flex-shrink-0"></span>
                          <span>汽车出口42.5万辆，同比增长18.9%</span>
                        </li>
                      </ul>
                    </div>

                    <p className="indent-8">
                      中国汽车工业协会表示，2026年汽车市场将继续保持平稳增长态势，新能源汽车仍将是推动市场增长的主要动力。协会将继续做好行业统计分析工作，为政府决策和企业经营提供数据支持。
                    </p>
                  </div>

                  {/* 责任编辑 */}
                  <div className="mt-12 pt-6 border-t-2 border-gray-200">
                    <p className="text-sm text-gray-600">
                      <span className="font-medium">责任编辑：</span>张明
                      <span className="mx-4">|</span>
                      <span className="font-medium">审核：</span>李华
                    </p>
                  </div>
                </div>
              </div>

              {/* 相关新闻 */}
              <div className="border-t border-gray-300 p-8 sm:p-12 bg-gray-50">
                <h3 className="text-xl font-bold text-gray-900 mb-6 pb-3 border-b-2 border-blue-800">
                  相关新闻
                </h3>
                <div className="space-y-4">
                  {relatedNews.map((news) => (
                    <a
                      key={news.id}
                      href={`#news-${news.id}`}
                      className="flex items-start gap-4 group"
                    >
                      <span className="w-2 h-2 bg-blue-800 rounded-full mt-2 flex-shrink-0"></span>
                      <div className="flex-1">
                        <h4 className="text-gray-900 group-hover:text-blue-800 transition-colors mb-1">
                          {news.title}
                        </h4>
                        <p className="text-sm text-gray-500">{news.date}</p>
                      </div>
                    </a>
                  ))}
                </div>
              </div>

              {/* 底部导航 */}
              <div className="border-t border-gray-300 p-6 flex justify-between">
                <a
                  href="#news"
                  className="flex items-center gap-2 text-blue-800 hover:text-red-600 transition-colors"
                >
                  <ArrowLeft className="w-4 h-4" />
                  返回列表
                </a>
                <div className="flex gap-4">
                  <a href="#prev" className="text-gray-600 hover:text-blue-800 transition-colors">
                    上一篇
                  </a>
                  <span className="text-gray-300">|</span>
                  <a href="#next" className="text-gray-600 hover:text-blue-800 transition-colors">
                    下一篇
                  </a>
                </div>
              </div>
            </div>
          </div>

          {/* 侧边栏 */}
          <div className="lg:col-span-1">
            {/* 热门新闻 */}
            <div className="bg-white border border-gray-300 mb-6">
              <div className="bg-blue-800 text-white px-4 py-3 font-medium">
                热门新闻
              </div>
              <div className="divide-y divide-gray-200">
                {hotNews.map((news, index) => (
                  <a
                    key={index}
                    href={`#hot-${index}`}
                    className="flex items-start gap-3 p-4 hover:bg-gray-50 transition-colors group"
                  >
                    <span className="flex-shrink-0 w-6 h-6 bg-red-600 text-white flex items-center justify-center text-sm font-bold">
                      {index + 1}
                    </span>
                    <div className="flex-1 min-w-0">
                      <h4 className="text-sm text-gray-900 group-hover:text-blue-800 transition-colors mb-1 line-clamp-2">
                        {news.title}
                      </h4>
                      <div className="flex items-center gap-1 text-xs text-gray-500">
                        <Eye className="w-3 h-3" />
                        {news.views}
                      </div>
                    </div>
                  </a>
                ))}
              </div>
            </div>

            {/* 最新通知 */}
            <div className="bg-white border border-gray-300 mb-6">
              <div className="bg-red-600 text-white px-4 py-3 font-medium">
                最新通知
              </div>
              <div className="divide-y divide-gray-200">
                {[
                  { title: "关于召开年度会议的通知", date: "01-06" },
                  { title: "会员企业调研通知", date: "01-05" },
                  { title: "标准征集公告", date: "01-04" },
                ].map((notice, index) => (
                  <a
                    key={index}
                    href={`#notice-${index}`}
                    className="block p-4 hover:bg-gray-50 transition-colors group"
                  >
                    <h4 className="text-sm text-gray-900 group-hover:text-blue-800 transition-colors mb-2 line-clamp-2">
                      {notice.title}
                    </h4>
                    <p className="text-xs text-gray-500">{notice.date}</p>
                  </a>
                ))}
              </div>
            </div>

            {/* 专题推荐 */}
            <div className="bg-white border border-gray-300">
              <div className="bg-blue-800 text-white px-4 py-3 font-medium">
                专题推荐
              </div>
              <div className="p-4 space-y-4">
                {[
                  { title: "碳达峰碳中和", color: "bg-green-700" },
                  { title: "智能网联汽车", color: "bg-blue-700" },
                  { title: "新能源汽车", color: "bg-cyan-700" },
                ].map((topic, index) => (
                  <a
                    key={index}
                    href={`#topic-${index}`}
                    className="block"
                  >
                    <div className={`${topic.color} text-white p-4 hover:opacity-90 transition-opacity`}>
                      <h4 className="font-bold">{topic.title}</h4>
                    </div>
                  </a>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
