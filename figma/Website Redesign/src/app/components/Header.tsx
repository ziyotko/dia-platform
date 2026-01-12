import { Menu, X, Search } from "lucide-react";
import { useState } from "react";

export function Header() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const navItems = [
    { name: "首页", href: "#home" },
    { name: "协会介绍", href: "#about" },
    { name: "新闻动态", href: "#news" },
    { name: "行业数据", href: "#stats" },
    { name: "会员服务", href: "#services" },
    { name: "政策法规", href: "#policy" },
    { name: "联系我们", href: "#contact" },
  ];

  return (
    <header className="fixed top-0 left-0 right-0 bg-white border-b-2 border-red-600 z-50">
      <div className="max-w-[1400px] mx-auto px-4 sm:px-6 lg:px-8">
        {/* Top Bar */}
        <div className="border-b border-gray-200 py-2">
          <div className="flex items-center justify-between text-xs">
            <div className="text-gray-600">
              <span>今天是：2026年1月7日 星期三</span>
            </div>
            <div className="flex items-center gap-4">
              <a href="#en" className="text-gray-600 hover:text-blue-800">English</a>
              <span className="text-gray-300">|</span>
              <a href="#contact" className="text-gray-600 hover:text-blue-800">联系我们</a>
              <span className="text-gray-300">|</span>
              <a href="#sitemap" className="text-gray-600 hover:text-blue-800">网站地图</a>
            </div>
          </div>
        </div>

        <div className="flex items-center justify-between h-24">
          {/* Logo */}
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 bg-blue-800 flex items-center justify-center">
              <span className="text-white text-2xl font-bold">CAAM</span>
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 tracking-wide">中国汽车工业协会</div>
              <div className="text-sm text-gray-600 tracking-wider">China Association of Automobile Manufacturers</div>
            </div>
          </div>

          {/* Desktop Navigation */}
          <nav className="hidden lg:flex items-center gap-1">
            {navItems.map((item) => (
              <a
                key={item.name}
                href={item.href}
                className="px-5 py-2 text-gray-700 hover:bg-blue-800 hover:text-white transition-colors"
              >
                {item.name}
              </a>
            ))}
          </nav>

          {/* Search and Mobile Menu Button */}
          <div className="flex items-center gap-4">
            <button className="p-2 text-gray-600 hover:text-blue-800 transition-colors">
              <Search className="w-5 h-5" />
            </button>
            <button
              className="lg:hidden p-2 text-gray-600"
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            >
              {mobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
            </button>
          </div>
        </div>

        {/* Mobile Menu */}
        {mobileMenuOpen && (
          <div className="lg:hidden py-4 border-t">
            <nav className="flex flex-col">
              {navItems.map((item) => (
                <a
                  key={item.name}
                  href={item.href}
                  className="text-gray-700 hover:bg-gray-50 py-3 px-4 border-b border-gray-100"
                  onClick={() => setMobileMenuOpen(false)}
                >
                  {item.name}
                </a>
              ))}
            </nav>
          </div>
        )}
      </div>
    </header>
  );
}