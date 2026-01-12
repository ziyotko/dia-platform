import { Mail, Phone, MapPin, Globe } from "lucide-react";

export function Footer() {
  const footerLinks = [
    {
      title: "关于协会",
      titleEn: "About Us",
      links: [
        { name: "协会简介", href: "#about" },
        { name: "组织架构", href: "#structure" },
        { name: "领导团队", href: "#leadership" },
        { name: "发展历程", href: "#history" },
        { name: "协会章程", href: "#charter" },
      ],
    },
    {
      title: "业务工作",
      titleEn: "Business",
      links: [
        { name: "政策研究", href: "#policy" },
        { name: "行业统计", href: "#statistics" },
        { name: "标准制定", href: "#standards" },
        { name: "国际合作", href: "#international" },
        { name: "行业自律", href: "#discipline" },
      ],
    },
    {
      title: "资讯服务",
      titleEn: "Information",
      links: [
        { name: "新闻动态", href: "#news" },
        { name: "通知公告", href: "#announcements" },
        { name: "行业报告", href: "#reports" },
        { name: "会议活动", href: "#events" },
        { name: "专题专栏", href: "#topics" },
      ],
    },
    {
      title: "会员中心",
      titleEn: "Members",
      links: [
        { name: "会员名录", href: "#members" },
        { name: "入会指南", href: "#guide" },
        { name: "会员服务", href: "#services" },
        { name: "会员登录", href: "#login" },
        { name: "在线申请", href: "#apply" },
      ],
    },
  ];

  return (
    <footer className="bg-gray-900 text-gray-300 border-t-4 border-red-600">
      <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
        {/* Main Footer Content */}
        <div className="py-12 grid sm:grid-cols-2 lg:grid-cols-5 gap-8">
          {/* Brand Section */}
          <div className="lg:col-span-1">
            <div className="w-16 h-16 bg-blue-800 flex items-center justify-center mb-4">
              <span className="text-white text-xl font-bold">CAAM</span>
            </div>
            <div className="text-white font-bold mb-2">中国汽车工业协会</div>
            <div className="text-xs text-gray-400 mb-6">
              China Association of<br />Automobile Manufacturers
            </div>
          </div>

          {/* Links Sections */}
          {footerLinks.map((section, index) => (
            <div key={index}>
              <h3 className="text-white font-bold mb-1">{section.title}</h3>
              <p className="text-xs text-gray-500 mb-4 pb-2 border-b border-gray-700">{section.titleEn}</p>
              <ul className="space-y-2">
                {section.links.map((link, linkIndex) => (
                  <li key={linkIndex}>
                    <a
                      href={link.href}
                      className="text-sm text-gray-400 hover:text-white transition-colors"
                    >
                      {link.name}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        {/* Contact Info */}
        <div className="border-t border-gray-800 py-8">
          <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-6">
            <div>
              <div className="text-sm font-bold text-white mb-2 flex items-center gap-2">
                <span className="w-1 h-4 bg-red-600"></span>
                地址 Address
              </div>
              <div className="text-sm text-gray-400">北京市西城区莲花池东路106号汇融大厦A座15层</div>
            </div>
            <div>
              <div className="text-sm font-bold text-white mb-2 flex items-center gap-2">
                <span className="w-1 h-4 bg-red-600"></span>
                联系电话 Tel
              </div>
              <div className="text-sm text-gray-400">010-63979900</div>
            </div>
            <div>
              <div className="text-sm font-bold text-white mb-2 flex items-center gap-2">
                <span className="w-1 h-4 bg-red-600"></span>
                电子邮箱 Email
              </div>
              <div className="text-sm text-gray-400">info@caam.org.cn</div>
            </div>
            <div>
              <div className="text-sm font-bold text-white mb-2 flex items-center gap-2">
                <span className="w-1 h-4 bg-red-600"></span>
                官方网站 Website
              </div>
              <div className="text-sm text-gray-400">www.caam.org.cn</div>
            </div>
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="border-t border-gray-800 py-6">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
            <div className="text-sm text-gray-400">
              版权所有 © 2026 中国汽车工业协会 京ICP备05030302号-1
            </div>
            <div className="flex gap-6 text-sm text-gray-400">
              <a href="#privacy" className="hover:text-white transition-colors">隐私政策</a>
              <span className="text-gray-700">|</span>
              <a href="#legal" className="hover:text-white transition-colors">法律声明</a>
              <span className="text-gray-700">|</span>
              <a href="#sitemap" className="hover:text-white transition-colors">网站地图</a>
            </div>
          </div>
        </div>
      </div>
    </footer>
  );
}