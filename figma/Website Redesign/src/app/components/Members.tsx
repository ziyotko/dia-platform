export function Members() {
  const memberCompanies = [
    { name: "一汽集团", logo: "一汽" },
    { name: "东风汽车", logo: "东风" },
    { name: "上汽集团", logo: "上汽" },
    { name: "长安汽车", logo: "长安" },
    { name: "北汽集团", logo: "北汽" },
    { name: "广汽集团", logo: "广汽" },
    { name: "比亚迪", logo: "比亚迪" },
    { name: "吉利控股", logo: "吉利" },
    { name: "长城汽车", logo: "长城" },
    { name: "奇瑞汽车", logo: "奇瑞" },
    { name: "蔚来汽车", logo: "蔚来" },
    { name: "小鹏汽车", logo: "小鹏" },
  ];

  return (
    <section className="py-16 bg-gray-50">
      <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
        <div className="border-l-4 border-red-600 pl-6 mb-12">
          <h2 className="text-3xl font-bold text-gray-900 mb-2">
            会员单位
          </h2>
          <p className="text-gray-600">Member Companies</p>
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-px bg-gray-300 border border-gray-300">
          {memberCompanies.map((company, index) => (
            <div
              key={index}
              className="bg-white p-6 flex items-center justify-center hover:bg-blue-50 transition-colors cursor-pointer"
            >
              <div className="text-center">
                <div className="w-16 h-16 bg-gradient-to-br from-blue-700 to-blue-900 mx-auto mb-3 flex items-center justify-center">
                  <span className="text-white font-bold text-sm">{company.logo}</span>
                </div>
                <div className="text-xs text-gray-700 font-medium">{company.name}</div>
              </div>
            </div>
          ))}
        </div>

        <div className="mt-8 text-center">
          <a
            href="#all-members"
            className="inline-block px-8 py-3 bg-blue-800 text-white hover:bg-blue-900 transition-colors"
          >
            查看全部会员单位
          </a>
        </div>
      </div>
    </section>
  );
}
