import { TrendingUp } from "lucide-react";

export function Stats() {
  const stats = [
    {
      value: "2,901万辆",
      label: "2025年汽车产量",
      sublabel: "Annual Production",
      trend: "+3.2%",
    },
    {
      value: "3.5万亿",
      label: "行业总产值（元）",
      sublabel: "Total Output Value",
      trend: "+5.8%",
    },
    {
      value: "3,000+",
      label: "会员企业数量",
      sublabel: "Member Companies",
      trend: "+12%",
    },
    {
      value: "35%",
      label: "新能源车占比",
      sublabel: "NEV Market Share",
      trend: "+8%",
    },
  ];

  return (
    <section id="stats" className="py-16 bg-gray-50">
      <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
        <div className="border-l-4 border-red-600 pl-6 mb-12">
          <h2 className="text-3xl font-bold text-gray-900 mb-2">
            行业数据统计
          </h2>
          <p className="text-gray-600">
            Industry Statistics
          </p>
        </div>

        <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-px bg-gray-300 border border-gray-300">
          {stats.map((stat, index) => (
            <div
              key={index}
              className="bg-white p-8 text-center"
            >
              <div className="text-4xl font-bold text-blue-800 mb-3">
                {stat.value}
              </div>
              <div className="text-sm text-gray-700 mb-1 font-medium">{stat.label}</div>
              <div className="text-xs text-gray-500 mb-3">{stat.sublabel}</div>
              <div className="inline-flex items-center gap-1 text-green-700 text-sm px-3 py-1 bg-green-50 border border-green-200">
                <span>同比 {stat.trend}</span>
              </div>
            </div>
          ))}
        </div>

        {/* Data Tables */}
        <div className="mt-12 grid lg:grid-cols-2 gap-8">
          <div className="bg-white border border-gray-300">
            <div className="bg-blue-800 text-white px-6 py-4 font-medium">
              月度产销数据
            </div>
            <div className="p-6">
              <table className="w-full">
                <thead>
                  <tr className="border-b-2 border-gray-300">
                    <th className="text-left py-3 text-gray-700">月份</th>
                    <th className="text-right py-3 text-gray-700">产量（万辆）</th>
                    <th className="text-right py-3 text-gray-700">销量（万辆）</th>
                  </tr>
                </thead>
                <tbody className="text-sm">
                  <tr className="border-b border-gray-200">
                    <td className="py-3 text-gray-700">2026年1月</td>
                    <td className="text-right py-3 text-gray-900">245.2</td>
                    <td className="text-right py-3 text-gray-900">242.8</td>
                  </tr>
                  <tr className="border-b border-gray-200">
                    <td className="py-3 text-gray-700">2025年12月</td>
                    <td className="text-right py-3 text-gray-900">278.6</td>
                    <td className="text-right py-3 text-gray-900">285.3</td>
                  </tr>
                  <tr className="border-b border-gray-200">
                    <td className="py-3 text-gray-700">2025年11月</td>
                    <td className="text-right py-3 text-gray-900">256.4</td>
                    <td className="text-right py-3 text-gray-900">261.7</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div className="bg-white border border-gray-300">
            <div className="bg-blue-800 text-white px-6 py-4 font-medium">
              新能源汽车数据
            </div>
            <div className="p-6">
              <table className="w-full">
                <thead>
                  <tr className="border-b-2 border-gray-300">
                    <th className="text-left py-3 text-gray-700">车型</th>
                    <th className="text-right py-3 text-gray-700">产量（万辆）</th>
                    <th className="text-right py-3 text-gray-700">占比</th>
                  </tr>
                </thead>
                <tbody className="text-sm">
                  <tr className="border-b border-gray-200">
                    <td className="py-3 text-gray-700">纯电动</td>
                    <td className="text-right py-3 text-gray-900">186.5</td>
                    <td className="text-right py-3 text-gray-900">76%</td>
                  </tr>
                  <tr className="border-b border-gray-200">
                    <td className="py-3 text-gray-700">插电混动</td>
                    <td className="text-right py-3 text-gray-900">51.3</td>
                    <td className="text-right py-3 text-gray-900">21%</td>
                  </tr>
                  <tr className="border-b border-gray-200">
                    <td className="py-3 text-gray-700">燃料电池</td>
                    <td className="text-right py-3 text-gray-900">7.4</td>
                    <td className="text-right py-3 text-gray-900">3%</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}