import { Header } from "./components/Header";
import { Hero } from "./components/Hero";
import { Stats } from "./components/Stats";
import { News } from "./components/News";
import { Announcements } from "./components/Announcements";
import { Services } from "./components/Services";
import { Members } from "./components/Members";
import { Policies } from "./components/Policies";
import { Reports } from "./components/Reports";
import { Topics } from "./components/Topics";
import { PartyBuilding } from "./components/PartyBuilding";
import { Links } from "./components/Links";
import { Footer } from "./components/Footer";
import { NewsDetail } from "./components/NewsDetail";
import { useState } from "react";

export default function App() {
  const [currentPage, setCurrentPage] = useState<'home' | 'newsDetail'>('home');

  return (
    <div className="min-h-screen bg-white">
      <Header />
      {currentPage === 'home' ? (
        <main>
          <Hero />
          <Stats />
          <News />
          <Announcements />
          <Policies />
          <Services />
          <Reports />
          <Members />
          <Topics />
          <PartyBuilding />
          <Links />
        </main>
      ) : (
        <NewsDetail />
      )}
      <Footer />
      
      {/* 页面切换按钮 - 仅用于演示 */}
      <div className="fixed bottom-8 right-8 z-50">
        <button
          onClick={() => setCurrentPage(currentPage === 'home' ? 'newsDetail' : 'home')}
          className="px-6 py-3 bg-blue-800 text-white shadow-lg hover:bg-blue-900 transition-colors"
        >
          {currentPage === 'home' ? '查看新闻详情' : '返回首页'}
        </button>
      </div>
    </div>
  );
}