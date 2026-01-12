import { Calendar, ArrowRight } from "lucide-react";

export function News() {
  const newsItems = [
    {
      id: 1,
      category: "协会动态",
      categoryEn: "Association News",
      title: "2026年1月汽车工业经济运行情况新闻发布会",
      description: "2026年1月，汽车产销分别完成245.2万辆和242.8万辆，环比分别下降11.9%和14.9%，同比分别增长3.5%和4.2%。",
      date: "2026-01-06",
      image: "https://images.unsplash.com/photo-1647427060118-4911c9821b82?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxjYXIlMjBtYW51ZmFjdHVyaW5nJTIwYXNzZW1ibHl8ZW58MXx8fHwxNzY3NzU2MzEzfDA&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral",
    },
    {
      id: 2,
      category: "政策解读",
      categoryEn: "Policy Interpretation",
      title: "关于新能源汽车推广应用推荐车型目录的公示",
      description: "根据《新能源汽车推广应用推荐车型目录管理办法》，现将申报的新能源汽车推广应用推荐车型目录进行公示。",
      date: "2026-01-05",
      image: "https://images.unsplash.com/photo-1593941707874-ef25b8b4a92b?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxlbGVjdHJpYyUyMHZlaGljbGUlMjBjaGFyZ2luZ3xlbnwxfHx8fDE3Njc3MzI1MTR8MA&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral",
    },
    {
      id: 3,
      category: "会议通知",
      categoryEn: "Meeting Notice",
      title: "关于召开2026年度中国汽车工业协会年会的通知",
      description: "定于2026年3月在北京召开2026年度中国汽车工业协会年会，现将有关事项通知如下。",
      date: "2026-01-03",
      image: "https://images.unsplash.com/photo-1689942007858-7b12bf5864bd?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxhdXRvbW90aXZlJTIwaW5kdXN0cnklMjB0ZWNobm9sb2d5fGVufDF8fHx8MTc2Nzc1NjMxM3ww&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral",
    },
  ];

  return (
    <section id="news" className="py-16 bg-white">
      <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between mb-12">
          <div className="border-l-4 border-red-600 pl-6">
            <h2 className="text-3xl font-bold text-gray-900 mb-2">
              新闻动态
            </h2>
            <p className="text-gray-600">News & Events</p>
          </div>
          <a
            href="#more-news"
            className="hidden sm:flex items-center text-blue-800 hover:text-red-600 transition-colors border border-gray-300 px-6 py-2 hover:border-red-600"
          >
            更多新闻
            <ArrowRight className="ml-2 w-4 h-4" />
          </a>
        </div>

        <div className="space-y-6">
          {newsItems.map((news) => (
            <article
              key={news.id}
              className="bg-white border border-gray-300 hover:shadow-md transition-shadow group cursor-pointer"
            >
              <div className="grid md:grid-cols-4 gap-0">
                <div className="md:col-span-1 aspect-video md:aspect-auto overflow-hidden">
                  <img
                    src={news.image}
                    alt={news.title}
                    className="w-full h-full object-cover"
                  />
                </div>
                <div className="md:col-span-3 p-6">
                  <div className="flex items-center gap-4 mb-3">
                    <div className="bg-red-600 text-white px-4 py-1">
                      <div className="text-xs font-medium">{news.category}</div>
                    </div>
                    <div className="text-gray-500 text-sm">
                      {news.date}
                    </div>
                  </div>
                  <h3 className="text-xl font-bold text-gray-900 mb-3 group-hover:text-blue-800 transition-colors">
                    {news.title}
                  </h3>
                  <p className="text-gray-600 leading-relaxed">
                    {news.description}
                  </p>
                </div>
              </div>
            </article>
          ))}
        </div>

        <div className="mt-8 text-center sm:hidden">
          <a
            href="#more-news"
            className="inline-flex items-center text-blue-800 hover:text-red-600 transition-colors border border-gray-300 px-6 py-2"
          >
            更多新闻
            <ArrowRight className="ml-2 w-4 h-4" />
          </a>
        </div>
      </div>
    </section>
  );
}