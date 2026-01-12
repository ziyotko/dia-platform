import { ChevronRight } from "lucide-react";

export function Hero() {
  return (
    <section id="home" className="relative pt-32 bg-white">
      {/* Red Top Border */}
      <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-red-600 via-red-500 to-red-600"></div>
      
      <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {/* Main Banner */}
        <div className="relative h-[500px] bg-gradient-to-r from-blue-900 to-blue-800 overflow-hidden">
          <img
            src="https://images.unsplash.com/photo-1585815302303-fcc8090ba1f5?crop=entropy&cs=tinysrgb&fit=max&fm=jpg&ixid=M3w3Nzg4Nzd8MHwxfHNlYXJjaHwxfHxtb2Rlcm4lMjBhdXRvbW9iaWxlJTIwZmFjdG9yeXxlbnwxfHx8fDE3Njc3NTYzMTJ8MA&ixlib=rb-4.1.0&q=80&w=1080&utm_source=figma&utm_medium=referral"
            alt="Modern automobile factory"
            className="absolute inset-0 w-full h-full object-cover opacity-40"
          />
          <div className="absolute inset-0 bg-gradient-to-r from-blue-900/90 to-blue-800/80"></div>
          
          <div className="relative h-full flex items-center">
            <div className="max-w-3xl px-12">
              <h1 className="text-5xl font-bold text-white mb-6 leading-tight">
                推动中国汽车工业高质量发展
              </h1>
              <p className="text-xl text-blue-100 mb-8 leading-relaxed">
                中国汽车工业协会是由中国汽车整车和零部件制造企业、汽车相关企业、团体及社会组织自愿组成的全国性行业组织
              </p>
              <div className="flex gap-4">
                <a
                  href="#about"
                  className="px-8 py-3 bg-red-600 text-white hover:bg-red-700 transition-colors"
                >
                  协会介绍
                </a>
                <a
                  href="#services"
                  className="px-8 py-3 bg-white text-blue-900 hover:bg-gray-100 transition-colors"
                >
                  会员服务
                </a>
              </div>
            </div>
          </div>
        </div>

        {/* Quick Links */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-px bg-gray-300 mt-8 border border-gray-300">
          {[
            { title: "行业统计", subtitle: "Industry Statistics" },
            { title: "政策法规", subtitle: "Policies & Regulations" },
            { title: "标准制定", subtitle: "Standard Development" },
            { title: "国际合作", subtitle: "International Cooperation" },
          ].map((item, index) => (
            <a
              key={index}
              href={`#${item.title}`}
              className="bg-white p-6 hover:bg-blue-50 transition-colors text-center border-l-4 border-transparent hover:border-blue-800"
            >
              <div className="text-lg font-bold text-gray-900 mb-1">{item.title}</div>
              <div className="text-xs text-gray-500">{item.subtitle}</div>
            </a>
          ))}
        </div>
      </div>
    </section>
  );
}