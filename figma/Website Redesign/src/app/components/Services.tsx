export function Services() {
  const services = [
    {
      number: "01",
      title: "政策研究与咨询",
      subtitle: "Policy Research & Consulting",
      description: "提供汽车产业政策解读、产业发展规划研究、标准制定参与等专业咨询服务。协助会员企业了解和应对国家产业政策。",
    },
    {
      number: "02",
      title: "行业统计与分析",
      subtitle: "Industry Statistics & Analysis",
      description: "建立完善的行业统计体系，定期发布产销数据、市场分析报告。为政府决策和企业经营提供数据支持。",
    },
    {
      number: "03",
      title: "标准制定与推广",
      subtitle: "Standard Development & Promotion",
      description: "组织制定汽车行业标准，参与国家标准、国际标准的制修订工作。推动行业技术进步和规范发展。",
    },
    {
      number: "04",
      title: "会员交流与培训",
      subtitle: "Member Exchange & Training",
      description: "组织行业交流活动、技术研讨会、专业培训。搭建会员企业沟通交流平台，促进资源共享和合作。",
    },
    {
      number: "05",
      title: "国际交流与合作",
      subtitle: "International Exchange & Cooperation",
      description: "开展国际交流活动，促进与国际汽车组织的合作。协助会员企业拓展海外市场，推动国际化发展。",
    },
    {
      number: "06",
      title: "行业自律与维权",
      subtitle: "Industry Discipline & Rights Protection",
      description: "制定行业自律规范，维护公平竞争环境。代表会员企业向政府反映诉求，维护行业和会员合法权益。",
    },
  ];

  return (
    <section id="services" className="py-16 bg-gray-50">
      <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
        <div className="border-l-4 border-red-600 pl-6 mb-12">
          <h2 className="text-3xl font-bold text-gray-900 mb-2">
            会员服务
          </h2>
          <p className="text-gray-600">
            Member Services
          </p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-px bg-gray-300 border border-gray-300">
          {services.map((service, index) => (
            <div
              key={index}
              className="bg-white p-8 hover:bg-blue-50 transition-colors"
            >
              <div className="flex items-baseline gap-3 mb-4 pb-4 border-b-2 border-blue-800">
                <div className="text-4xl font-bold text-blue-800">
                  {service.number}
                </div>
                <div className="flex-1">
                  <h3 className="text-lg font-bold text-gray-900 mb-1">
                    {service.title}
                  </h3>
                  <p className="text-xs text-gray-500">{service.subtitle}</p>
                </div>
              </div>
              <p className="text-gray-600 leading-relaxed text-sm">
                {service.description}
              </p>
            </div>
          ))}
        </div>

        {/* CTA Section */}
        <div className="mt-12 bg-blue-800 border-4 border-red-600">
          <div className="p-12 text-center text-white">
            <h3 className="text-3xl font-bold mb-4">
              申请加入中国汽车工业协会
            </h3>
            <p className="text-blue-100 mb-8 max-w-3xl mx-auto leading-relaxed">
              中国汽车工业协会诚邀汽车整车和零部件制造企业、汽车相关企业及组织加入。
              成为会员，您将享受全方位的专业服务，参与行业标准制定，拓展行业资源网络。
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <a
                href="#membership"
                className="inline-flex items-center justify-center px-10 py-4 bg-red-600 text-white hover:bg-red-700 transition-colors"
              >
                入会申请
              </a>
              <a
                href="#contact"
                className="inline-flex items-center justify-center px-10 py-4 bg-white text-blue-900 hover:bg-gray-100 transition-colors"
              >
                咨询了解
              </a>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
