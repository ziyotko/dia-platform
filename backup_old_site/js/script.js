$(document).ready(function() {
    // ==========================================
    // 1. Data Source (Mock API Response)
    // ==========================================
    const mockData = {
        notifications: [
            '关于推荐评选第二十六届中国专利奖的通知',
            '关于召开2025汽车工业统计年报工作会议的通知',
            '关于开展汽车行业反歧视调查的通知',
            '关于2025智能汽车基础软件生态大会时间变更的通知'
        ],
        industryNews: [
            { id: 1, title: '2025年1月汽车工业经济运行情况', date: '2025-02-10', thumb: 'https://picsum.photos/seed/n1/240/160', excerpt: '1月，汽车产销分别完成241万辆和243.9万辆，同比分别增长51.2%和47.9%。' },
            { id: 2, title: '五部门发布关于开展智能网联汽车“车路云一体化”应用试点工作的通知', date: '2025-01-18', thumb: 'https://picsum.photos/seed/n2/240/160', excerpt: '工业和信息化部、公安部、自然资源部、住房和城乡建设部、交通运输部...' },
            { id: 3, title: '中汽协会发布2024年汽车出口情况简析', date: '2025-01-15', thumb: 'https://picsum.photos/seed/n3/240/160', excerpt: '据中国汽车工业协会分析，2024年12月，汽车出口49.9万辆，环比增长3.5%。' }
        ],
        workNews: [
            { id: 1, title: '中汽协会组织召开汽车行业标准化工作会议', date: '2025-12-20' },
            { id: 2, title: '付炳锋常务副会长会见德国汽车工业协会代表', date: '2025-12-18' },
            { id: 3, title: '许海东副总工程师出席新能源汽车发展论坛', date: '2025-12-15' },
            { id: 4, title: '中汽协会发布2025年汽车市场预测报告', date: '2025-12-10' },
            { id: 5, title: '第13届中国汽车论坛组委会召开第一次会议', date: '2025-12-05' }
        ],
        announcements: [
            { id: 1, title: '关于发布团体标准《乘用车空气动力学性能术语》的公告', date: '2025-01-05' },
            { id: 2, title: '关于召开2025中国汽车供应链大会的通知', date: '2025-01-03' },
            { id: 3, title: '关于开展2024年度汽车行业科学技术奖推荐工作的通知', date: '2024-12-28' },
            { id: 4, title: '中汽协会关于吸纳新会员的公告（2024年第4批）', date: '2024-12-25' },
            { id: 5, title: '关于举办新能源汽车下乡活动的通知', date: '2024-12-20' }
        ],
        services: [
            { key: 'member', title: '会员专区', desc: '入会申请与会员服务' },
            { key: 'cert', title: '合格证查询', desc: '车辆合格证信息查询' },
            { key: 'standard', title: '标准法规', desc: '行业标准检索与下载' },
            { key: 'announce', title: '公告查询', desc: '企业及产品公告查询' },
            { key: 'parts', title: '零部件目录', desc: '零部件产品目录查询' },
            { key: 'credit', title: '信用评价', desc: '行业信用信息查询' }
        ],
        topics: [
            { title: '电动·智能', link: '#' },
            { title: '政策法规', link: '#' },
            { title: '统计数据', link: '#' },
            { title: '协会概况', link: '#' },
            { title: '行业动态', link: '#' },
            { title: '会员专区', link: '#' }
        ],
        subsites: [
            { category: '分支机构', items: ['整车分会', '乘用车分会', '商用车分会', '零部件分会'] },
            { category: '专题子站', items: ['智能汽车', '汽车芯片', '再制造', '后市场'] }
        ],
        stats: [
            { label: '汽车产销', value: '3000万+', unit: '辆', change: '+12%' },
            { label: '新能源', value: '950万', unit: '辆', change: '+35%' },
            { label: '出口量', value: '480万', unit: '辆', change: '+55%' }
        ],
        policies: [
            '关于汽车数据处理5项安全要求检测情况的通报',
            '工业和信息化部关于印发《国家汽车芯片标准体系建设指南》的通知',
            '关于开展智能网联汽车准入和上路通行试点工作的通知',
            '四部门关于开展智能网联汽车准入和上路通行试点工作的通知'
        ]
    };

    // Icon helper (SVG paths)
    const icons = {
        member: '<path d="M4 7h16v2H4V7zm0 4h12v2H4v-2zm0 4h8v2H4v-2z" fill="currentColor"/>',
        cert: '<path d="M12 2l3 6 6 .9-4.5 4.2 1.1 6.7L12 17l-5.6 2.8 1.1-6.7L3 8.9 9 8l3-6z" fill="currentColor"/>',
        standard: '<path d="M4 4h12v2H6v12H4V4zm6 4h8v12H10V8zm2 2h4v2h-4v-2zm0 4h4v2h-4v-2z" fill="currentColor"/>',
        announce: '<path d="M3 10h2l3-3h7l3-3v14l-3-3H8l-3-3H3v-2z" fill="currentColor"/>',
        parts: '<path d="M12 2a4 4 0 110 8 4 4 0 010-8zm-7 18a7 7 0 0114 0H5z" fill="currentColor"/>',
        credit: '<path d="M3 5h18v4H3V5zm0 6h12v8H3v-8zm14 2h4v6h-4v-6z" fill="currentColor"/>',
        default: '<circle cx="12" cy="12" r="10" fill="currentColor"/>'
    };

    // ==========================================
    // 2. Rendering Logic
    // ==========================================

    // Render Notifications
    const $notifyWrapper = $('.notify-wrapper');
    mockData.notifications.forEach(note => {
        $notifyWrapper.append(`<span class="notify-item">${note}</span>`);
    });

    // Render Industry News
    const $industryNewsList = $('#industry-news-list');
    mockData.industryNews.forEach(item => {
        $industryNewsList.append(`
            <div class="news-card">
                <img src="${item.thumb}" alt="${item.title}" class="news-thumb">
                <div class="news-info">
                    <div class="news-title">${item.title}</div>
                    <div class="news-excerpt">${item.excerpt}</div>
                    <div class="news-date">${item.date}</div>
                </div>
            </div>
        `);
    });

    // Render Work News & Announcements (List Style)
    function renderList(targetId, data) {
        const $target = $(targetId);
        data.forEach(item => {
            $target.append(`
                <li>
                    <a href="#" class="title" title="${item.title}">${item.title}</a>
                    <span class="date">${item.date}</span>
                </li>
            `);
        });
    }
    renderList('#work-news-list', mockData.workNews);
    renderList('#announcement-list', mockData.announcements);

    // Render Services
    const $servicesList = $('#services-list');
    mockData.services.forEach(s => {
        const iconSvg = `<svg viewBox="0 0 24 24" width="24" height="24" class="service-icon">${icons[s.key] || icons.default}</svg>`;
        $servicesList.append(`
            <div class="service-card">
                ${iconSvg}
                <div class="service-info">
                    <h3>${s.title}</h3>
                    <p>${s.desc}</p>
                </div>
            </div>
        `);
    });

    // Render Topics
    const $topicsList = $('#topics-list');
    mockData.topics.forEach(t => {
        $topicsList.append(`
            <a href="${t.link}" class="topic-card">
                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle></svg>
                ${t.title}
            </a>
        `);
    });

    // Render Subsites
    const $subsList = $('#subs-list');
    mockData.subsites.forEach(cat => {
        const itemsHtml = cat.items.map(i => `<a href="#" class="subs-chip">${i}</a>`).join('');
        $subsList.append(`
            <div class="subs-category">
                <div class="subs-title">${cat.category}</div>
                <div class="subs-items">${itemsHtml}</div>
            </div>
        `);
    });

    // Render Stats Data
    const $statsData = $('#stats-data');
    mockData.stats.forEach(s => {
        $statsData.append(`
            <div class="stat-item">
                <div class="stat-value">${s.value} <small style="font-size:12px; color:#666">${s.unit}</small></div>
                <div class="stat-label">${s.label}</div>
                <div class="stat-change">${s.change}</div>
            </div>
        `);
    });

    // Render Simple Chart
    const $chart = $('#stats-chart');
    mockData.stats.forEach((s, i) => {
        const height = i === 0 ? 80 : i === 1 ? 60 : 40; // Mock heights
        const left = i * 33 + 10;
        $chart.append(`<div class="chart-bar" style="height: ${height}%; left: ${left}%; bottom: 0; width: 20%;"></div>`);
    });

    // Render Policies
    const $policyList = $('#policy-list');
    mockData.policies.forEach(p => {
        $policyList.append(`
            <div class="policy-card">
                <span class="policy-icon">⚖️</span>
                <p style="font-size:14px; color:#333; display:-webkit-box; -webkit-line-clamp:2; -webkit-box-orient:vertical; overflow:hidden;">${p}</p>
            </div>
        `);
    });

    // ==========================================
    // 3. Carousel Logic
    // ==========================================
    let currentSlide = 0;
    const $slides = $('.carousel-slide');
    const $indicators = $('.indicator');
    const totalSlides = $slides.length;

    function showSlide(index) {
        $slides.removeClass('active').css('opacity', 0);
        $indicators.removeClass('active');
        
        currentSlide = (index + totalSlides) % totalSlides;
        
        $slides.eq(currentSlide).addClass('active').css('opacity', 1);
        $indicators.eq(currentSlide).addClass('active');
    }

    $('.carousel-btn.next').click(() => showSlide(currentSlide + 1));
    $('.carousel-btn.prev').click(() => showSlide(currentSlide - 1));
    $indicators.click(function() {
        showSlide($(this).data('slide'));
    });

    // Auto play
    let slideInterval = setInterval(() => showSlide(currentSlide + 1), 5000);
    $('.carousel-container').hover(
        () => clearInterval(slideInterval),
        () => slideInterval = setInterval(() => showSlide(currentSlide + 1), 5000)
    );

    // ==========================================
    // 4. Notification Scrolling
    // ==========================================
    let notifyIndex = 0;
    const $notifyItems = $('.notify-item'); // Note: This captures items AFTER render
    // Simple vertical scroll effect
    setInterval(() => {
        const itemHeight = 24;
        notifyIndex = (notifyIndex + 1) % mockData.notifications.length;
        $notifyWrapper.css('transition', 'transform 0.5s').css('transform', `translateY(-${notifyIndex * itemHeight}px)`);
    }, 3000);

});