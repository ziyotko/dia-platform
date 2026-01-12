import { ExternalLink } from "lucide-react";

export function Links() {
  const linkCategories = [
    {
      title: "政府部门",
      links: [
        "工业和信息化部",
        "国家发展改革委",
        "交通运输部",
        "科学技术部",
        "商务部",
        "生态环境部",
      ],
    },
    {
      title: "行业组织",
      links: [
        "中国机械工业联合会",
        "中国汽车工程学会",
        "中国汽车流通协会",
        "中国汽车技术研究中心",
        "中国质量认证中心",
        "中国汽车报社",
      ],
    },
    {
      title: "国际组织",
      links: [
        "国际汽车制造商协会",
        "世界汽车工程师学会联合会",
        "亚太汽车工业协会联合会",
        "日本汽车工业协会",
        "美国汽车工业协会",
        "欧洲汽车制造商协会",
      ],
    },
  ];

  return (
    <section className="py-16 bg-gray-50">
      <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
        <div className="border-l-4 border-red-600 pl-6 mb-12">
          <h2 className="text-3xl font-bold text-gray-900 mb-2">
            友情链接
          </h2>
          <p className="text-gray-600">Useful Links</p>
        </div>

        <div className="grid md:grid-cols-3 gap-8">
          {linkCategories.map((category, index) => (
            <div key={index} className="bg-white border border-gray-300">
              <div className="bg-blue-800 text-white px-5 py-3 font-medium">
                {category.title}
              </div>
              <div className="p-5">
                <ul className="space-y-3">
                  {category.links.map((link, linkIndex) => (
                    <li key={linkIndex}>
                      <a
                        href={`#link-${index}-${linkIndex}`}
                        className="flex items-center gap-2 text-gray-700 hover:text-blue-800 transition-colors text-sm group"
                      >
                        <ExternalLink className="w-3 h-3 text-gray-400 group-hover:text-blue-800" />
                        {link}
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
