import { Award } from "lucide-react";

export function PartyBuilding() {
  return (
    <section className="py-16 bg-red-50">
      <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid lg:grid-cols-3 gap-8">
          {/* 党建工作 */}
          <div className="lg:col-span-2">
            <div className="border-l-4 border-red-600 pl-6 mb-8">
              <h2 className="text-2xl font-bold text-gray-900 mb-2">
                党建工作
              </h2>
              <p className="text-gray-600">Party Building</p>
            </div>

            <div className="bg-white border border-gray-300 p-8">
              <div className="flex items-start gap-6">
                <div className="flex-shrink-0">
                  <div className="w-24 h-24 bg-red-600 flex items-center justify-center">
                    <Award className="w-12 h-12 text-yellow-400" />
                  </div>
                </div>
                <div className="flex-1">
                  <h3 className="text-xl font-bold text-gray-900 mb-4">
                    坚持党的领导 推动行业发展
                  </h3>
                  <div className="space-y-3 text-gray-700 leading-relaxed">
                    <p>
                      中国汽车工业协会党委认真学习贯彻习近平新时代中国特色社会主义思想，
                      坚持党对协会工作的全面领导，充分发挥党组织的政治核心作用。
                    </p>
                    <p>
                      协会党委积极开展党史学习教育，推动党建工作与业务工作深度融合，
                      引领广大党员干部在推动汽车产业高质量发展中担当作为。
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <div className="grid sm:grid-cols-2 gap-4 mt-6">
              {[
                { title: "学习贯彻党的二十大精神", date: "2026-01-05" },
                { title: "开展主题党日活动", date: "2026-01-03" },
              ].map((item, index) => (
                <a
                  key={index}
                  href={`#party-${index}`}
                  className="bg-white border border-gray-300 p-5 hover:shadow-md transition-shadow"
                >
                  <div className="text-xs text-red-600 mb-2">{item.date}</div>
                  <h4 className="font-bold text-gray-900 hover:text-blue-800 transition-colors">
                    {item.title}
                  </h4>
                </a>
              ))}
            </div>
          </div>

          {/* 行业荣誉 */}
          <div>
            <div className="border-l-4 border-red-600 pl-6 mb-8">
              <h2 className="text-2xl font-bold text-gray-900 mb-2">
                行业荣誉
              </h2>
              <p className="text-gray-600">Honors</p>
            </div>

            <div className="bg-white border border-gray-300">
              <div className="bg-red-600 text-white px-4 py-3 font-medium">
                获奖项目
              </div>
              <div className="divide-y divide-gray-200">
                {[
                  { year: "2025", title: "全国先进社会组织" },
                  { year: "2024", title: "中国汽车工业科学技术进步奖" },
                  { year: "2024", title: "行业标准化工作先进单位" },
                  { year: "2023", title: "全国性行业协会商会先进党组织" },
                ].map((honor, index) => (
                  <div key={index} className="px-4 py-4">
                    <div className="flex items-start gap-3">
                      <div className="w-12 h-12 bg-yellow-400 flex items-center justify-center flex-shrink-0">
                        <Award className="w-6 h-6 text-yellow-700" />
                      </div>
                      <div>
                        <div className="text-xs text-gray-500 mb-1">{honor.year}年度</div>
                        <div className="text-sm font-medium text-gray-900">{honor.title}</div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
