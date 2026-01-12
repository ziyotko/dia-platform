import { BookOpen, Download } from "lucide-react";

export function Reports() {
  const reports = [
    {
      id: 1,
      title: "中国汽车工业年鉴 2025",
      cover: "https://images.unsplash.com/photo-1689942007858-7b12bf5864bd?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxhdXRvbW90aXZlJTIwaW5kdXN0cnklMjB0ZWNobm9sb2d5fGVufDF8fHx8MTc2Nzc1NjMxM3ww&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral",
      description: "全面总结2025年度中国汽车工业发展情况",
      date: "2025-12",
    },
    {
      id: 2,
      title: "新能源汽车产业发展报告",
      cover: "https://images.unsplash.com/photo-1593941707874-ef25b8b4a92b?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxlbGVjdHJpYyUyMHZlaGljbGUlMjBjaGFyZ2luZ3xlbnwxfHx8fDE3Njc3MzI1MTR8MA&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral",
      description: "深度分析新能源汽车产业现状与趋势",
      date: "2025-11",
    },
    {
      id: 3,
      title: "智能网联汽车技术路线图",
      cover: "https://images.unsplash.com/photo-1647427060118-4911c9821b82?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxjYXIlMjBtYW51ZmFjdHVyaW5nJTIwYXNzZW1ibHl8ZW58MXx8fHwxNzY3NzU2MzEzfDA&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral",
      description: "智能网联汽车技术发展规划与实施路径",
      date: "2025-10",
    },
    {
      id: 4,
      title: "汽车产业链供应链白皮书",
      cover: "https://images.unsplash.com/photo-1585815302303-fcc8090ba1f5?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxtb2Rlcm4lMjBhdXRvbW9iaWxlJTIwZmFjdG9yeXxlbnwxfHx8fDE3Njc3NTYzMTJ8MA&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral",
      description: "汽车产业链供应链安全与韧性研究",
      date: "2025-09",
    },
  ];

  return (
    <section className="py-16 bg-gray-50">
      <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
        <div className="border-l-4 border-red-600 pl-6 mb-12">
          <h2 className="text-3xl font-bold text-gray-900 mb-2">
            行业报告
          </h2>
          <p className="text-gray-600">Industry Reports & Publications</p>
        </div>

        <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {reports.map((report) => (
            <div
              key={report.id}
              className="bg-white border border-gray-300 hover:shadow-lg transition-shadow group"
            >
              <div className="aspect-[3/4] overflow-hidden border-b border-gray-300">
                <img
                  src={report.cover}
                  alt={report.title}
                  className="w-full h-full object-cover"
                />
              </div>
              <div className="p-5">
                <div className="flex items-center gap-2 mb-3">
                  <BookOpen className="w-4 h-4 text-blue-800" />
                  <span className="text-xs text-gray-500">{report.date}</span>
                </div>
                <h3 className="text-base font-bold text-gray-900 mb-2 group-hover:text-blue-800 transition-colors">
                  {report.title}
                </h3>
                <p className="text-sm text-gray-600 mb-4">{report.description}</p>
                <button className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-blue-800 text-white hover:bg-blue-900 transition-colors text-sm">
                  <Download className="w-4 h-4" />
                  下载报告
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
