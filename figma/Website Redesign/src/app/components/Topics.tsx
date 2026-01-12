export function Topics() {
  const topics = [
    {
      id: 1,
      title: "碳达峰碳中和",
      subtitle: "Carbon Neutrality",
      description: "汽车产业绿色低碳转型发展",
      color: "from-green-700 to-green-900",
    },
    {
      id: 2,
      title: "智能网联汽车",
      subtitle: "Intelligent Connected Vehicles",
      description: "推动汽车智能化网联化发展",
      color: "from-blue-700 to-blue-900",
    },
    {
      id: 3,
      title: "新能源汽车",
      subtitle: "New Energy Vehicles",
      description: "促进新能源汽车产业高质量发展",
      color: "from-cyan-700 to-cyan-900",
    },
    {
      id: 4,
      title: "走出去战略",
      subtitle: "Going Global",
      description: "支持汽车企业国际化发展",
      color: "from-purple-700 to-purple-900",
    },
  ];

  return (
    <section className="py-16 bg-white">
      <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
        <div className="border-l-4 border-red-600 pl-6 mb-12">
          <h2 className="text-3xl font-bold text-gray-900 mb-2">
            专题专栏
          </h2>
          <p className="text-gray-600">Special Topics</p>
        </div>

        <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-px bg-gray-300 border border-gray-300">
          {topics.map((topic) => (
            <a
              key={topic.id}
              href={`#topic-${topic.id}`}
              className="group relative overflow-hidden bg-white aspect-[4/3]"
            >
              <div className={`absolute inset-0 bg-gradient-to-br ${topic.color} opacity-90 group-hover:opacity-100 transition-opacity`}></div>
              <div className="relative h-full flex flex-col items-center justify-center text-center p-6 text-white">
                <h3 className="text-2xl font-bold mb-2">{topic.title}</h3>
                <p className="text-sm text-blue-100 mb-3">{topic.subtitle}</p>
                <p className="text-sm">{topic.description}</p>
                <div className="mt-4 w-12 h-0.5 bg-white/50 group-hover:w-20 transition-all"></div>
              </div>
            </a>
          ))}
        </div>
      </div>
    </section>
  );
}
