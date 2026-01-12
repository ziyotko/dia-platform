import { FileText, Download } from "lucide-react";

export function Policies() {
  const policies = [
    {
      id: 1,
      category: "国家政策",
      title: "关于进一步构建高质量充电基础设施体系的指导意见",
      department: "国家发展改革委",
      date: "2025-12-28",
    },
    {
      id: 2,
      category: "行业标准",
      title: "汽车动力蓄电池编码规则（GB/T 34014-2023）",
      department: "国家市场监督管理总局",
      date: "2025-12-15",
    },
    {
      id: 3,
      category: "地方政策",
      title: "北京市关于促进汽车消费的若干措施",
      department: "北京市人民政府",
      date: "2025-12-10",
    },
    {
      id: 4,
      category: "国家政策",
      title: "智能网联汽车准入和上路通行试点工作的通知",
      department: "工业和信息化部",
      date: "2025-12-05",
    },
  ];

  return (
    <section className="py-16 bg-white">
      <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid lg:grid-cols-3 gap-8">
          {/* 政策法规 */}
          <div className="lg:col-span-2">
            <div className="border-l-4 border-red-600 pl-6 mb-8">
              <h2 className="text-2xl font-bold text-gray-900 mb-2">
                政策法规
              </h2>
              <p className="text-gray-600">Policies & Regulations</p>
            </div>

            <div className="space-y-4">
              {policies.map((policy) => (
                <div
                  key={policy.id}
                  className="bg-white border border-gray-300 p-5 hover:shadow-md transition-shadow"
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-3">
                        <span className="px-3 py-1 bg-blue-800 text-white text-xs">
                          {policy.category}
                        </span>
                        <span className="text-xs text-gray-500">{policy.department}</span>
                      </div>
                      <h3 className="text-base font-bold text-gray-900 mb-2 hover:text-blue-800 transition-colors">
                        {policy.title}
                      </h3>
                      <div className="text-xs text-gray-500">发布时间：{policy.date}</div>
                    </div>
                    <button className="p-2 border border-gray-300 hover:bg-gray-50 transition-colors">
                      <Download className="w-4 h-4 text-gray-600" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* 标准目录 */}
          <div>
            <div className="border-l-4 border-red-600 pl-6 mb-8">
              <h2 className="text-2xl font-bold text-gray-900 mb-2">
                标准目录
              </h2>
              <p className="text-gray-600">Standards</p>
            </div>

            <div className="bg-white border border-gray-300">
              <div className="bg-gray-100 px-4 py-3 border-b border-gray-300">
                <h3 className="font-bold text-gray-900 text-sm">最新标准</h3>
              </div>
              <div className="divide-y divide-gray-200">
                {[
                  { code: "GB/T 18488.1", name: "电动汽车用电机及其控制器" },
                  { code: "GB/T 31467.3", name: "电动汽车用锂离子动力蓄电池" },
                  { code: "GB 18352.6", name: "轻型汽车污染物排放限值" },
                  { code: "GB/T 19596", name: "电动汽车术语" },
                  { code: "GB/T 20234.1", name: "电动汽车传导充电用连接装置" },
                ].map((standard, index) => (
                  <a
                    key={index}
                    href={`#standard-${index}`}
                    className="block px-4 py-3 hover:bg-gray-50 transition-colors"
                  >
                    <div className="text-sm font-medium text-gray-900 mb-1">{standard.code}</div>
                    <div className="text-xs text-gray-600">{standard.name}</div>
                  </a>
                ))}
              </div>
              <div className="border-t border-gray-300 bg-gray-50 px-4 py-3 text-center">
                <a href="#all-standards" className="text-blue-800 hover:text-red-600 text-sm">
                  查看全部标准 →
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
