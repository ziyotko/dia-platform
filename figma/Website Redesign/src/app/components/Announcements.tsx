import { Bell, FileText } from "lucide-react";

export function Announcements() {
  const announcements = [
    {
      id: 1,
      type: "通知",
      title: "关于召开中国汽车工业协会第十届会员代表大会的通知",
      date: "2026-01-06",
    },
    {
      id: 2,
      type: "公告",
      title: "2026年度汽车行业标准制修订项目征集公告",
      date: "2026-01-05",
    },
    {
      id: 3,
      type: "通知",
      title: "关于开展2026年度优秀会员企业评选活动的通知",
      date: "2026-01-04",
    },
    {
      id: 4,
      type: "公示",
      title: "2025年度中国汽车工业科学技术奖拟授奖项目公示",
      date: "2026-01-03",
    },
    {
      id: 5,
      type: "通知",
      title: "关于举办汽车产业碳中和发展论坛的通知",
      date: "2026-01-02",
    },
  ];

  return (
    <section className="py-16 bg-white">
      <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
        <div className="border-l-4 border-red-600 pl-6 mb-12">
          <h2 className="text-3xl font-bold text-gray-900 mb-2">
            通知公告
          </h2>
          <p className="text-gray-600">Announcements & Notices</p>
        </div>

        <div className="bg-white border border-gray-300">
          <div className="bg-blue-800 text-white px-6 py-4 font-medium flex items-center gap-2">
            <Bell className="w-5 h-5" />
            最新公告
          </div>
          <div className="divide-y divide-gray-200">
            {announcements.map((item) => (
              <a
                key={item.id}
                href={`#announcement-${item.id}`}
                className="flex items-center justify-between px-6 py-4 hover:bg-gray-50 transition-colors group"
              >
                <div className="flex items-center gap-4 flex-1">
                  <FileText className="w-5 h-5 text-gray-400" />
                  <span className="px-3 py-1 bg-red-600 text-white text-xs">
                    {item.type}
                  </span>
                  <span className="text-gray-900 group-hover:text-blue-800 transition-colors">
                    {item.title}
                  </span>
                </div>
                <span className="text-gray-500 text-sm ml-4">{item.date}</span>
              </a>
            ))}
          </div>
          <div className="border-t border-gray-300 bg-gray-50 px-6 py-3 text-center">
            <a href="#more-announcements" className="text-blue-800 hover:text-red-600 text-sm">
              查看更多公告 →
            </a>
          </div>
        </div>
      </div>
    </section>
  );
}
