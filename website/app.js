// --- AeroBid 官方门户核心交互逻辑 ---

document.addEventListener('DOMContentLoaded', () => {
  initNavbar();
  initMobileMenu();
  initSDKTabs();
  initBiddingSimulator();
  initDockerCopy();
  initDocsModal();
});

// 1. 导航栏滚动效果
function initNavbar() {
  const navbar = document.getElementById('navbar');
  window.addEventListener('scroll', () => {
    if (window.scrollY > 50) {
      navbar.classList.add('navbar-scrolled');
    } else {
      navbar.classList.remove('navbar-scrolled');
    }
  });
}

// 2. 移动端菜单切换
function initMobileMenu() {
  const menuToggle = document.getElementById('menu-toggle');
  const navLinks = document.getElementById('nav-links');

  if (menuToggle && navLinks) {
    menuToggle.addEventListener('click', () => {
      navLinks.classList.toggle('active');
      const isExpanded = navLinks.classList.contains('active');
      menuToggle.innerHTML = isExpanded ? '✕' : '☰';
    });

    // 点击链接关闭菜单
    navLinks.querySelectorAll('a').forEach(link => {
      link.addEventListener('click', () => {
        navLinks.classList.remove('active');
        menuToggle.innerHTML = '☰';
      });
    });
  }
}

// 3. SDK 开发者集成代码切换
const sdkSnippets = {
  swift: `<span class="code-keyword">import</span> AeroBidSDK

<span class="code-comment">// 1. 初始化 SDK 核心</span>
AeroBidSDK.shared.initialize(
    appKey: <span class="code-string">"ab_prod_8f9c0e4"</span>,
    appSecret: <span class="code-string">"sec_7d2f9b8c"</span>
) { success <span class="code-keyword">in</span>
    print(<span class="code-string">"AeroBid initialize: \\(success)"</span>)
}

<span class="code-comment">// 2. 加载广告位对应素材</span>
<span class="code-keyword">let</span> banner = AeroBidBanner(placementId: <span class="code-string">"plc_home_banner"</span>)
banner.loadAd { [weak <span class="code-keyword">self</span>] result <span class="code-keyword">in</span>
    <span class="code-keyword">switch</span> result {
    <span class="code-keyword">case</span> .success(let ad):
        <span class="code-comment">// 3. 在视口中渲染</span>
        ad.show(in: <span class="code-keyword">self</span>?.view)
        print(<span class="code-string">"Ad loaded via: \\(ad.dspName), CPM: \\(ad.cpm)"</span>)
    <span class="code-keyword">case</span> .failure(let error):
        print(<span class="code-string">"Ad failed to load: \\(error.localizedDescription)"</span>)
    }
}`,

  kotlin: `<span class="code-keyword">import</span> com.aerobid.sdk.AeroBid
<span class="code-keyword">import</span> com.aerobid.sdk.ads.AeroBidInterstitial

<span class="code-comment">// 1. 初始化 SDK</span>
AeroBid.initialize(
    context = this,
    appKey = <span class="code-string">"ab_prod_8f9c0e4"</span>,
    appSecret = <span class="code-string">"sec_7d2f9b8c"</span>
)

<span class="code-comment">// 2. 载入插屏广告 (Interstitial)</span>
val interstitial = AeroBidInterstitial(this, <span class="code-string">"plc_detail_inter"</span>)
interstitial.load(object : AeroBidAdListener {
    override fun <span class="code-keyword">onAdLoaded</span>(ad: AeroBidAd) {
        <span class="code-comment">// 3. 展示广告并触发自动追踪回调</span>
        interstitial.show()
        Log.d(<span class="code-string">"AeroBid"</span>, <span class="code-string">"Ad filled: \${ad.cpm} USD, DSP: \${ad.sourceName}"</span>)
    }

    override fun <span class="code-keyword">onAdFailed</span>(error: AeroBidError) {
        Log.e(<span class="code-string">"AeroBid"</span>, <span class="code-string">"Ad failed to load: \${error.message}"</span>)
    }
})`,

  js: `<span class="code-keyword">import</span> { AeroBidEngine } <span class="code-keyword">from</span> <span class="code-string">'@aerobid/web-sdk'</span>;

<span class="code-comment">// 1. 建立 Web 竞价客户端</span>
<span class="code-keyword">const</span> client = <span class="code-keyword">new</span> AeroBidEngine({
  appKey: <span class="code-string">'ab_prod_8f9c0e4'</span>,
  endpoint: <span class="code-string">'https://api.aerobid.io/v1'</span>
});

<span class="code-comment">// 2. 发起广告竞价请求</span>
client.requestBid({
  placementId: <span class="code-string">'plc_web_sidebar'</span>,
  sizes: [[300, 250]]
}).then(ad => {
  <span class="code-comment">// 3. 渲染广告创意</span>
  document.getElementById(<span class="code-string">'ad-slot'</span>).innerHTML = ad.creativeHTML;
  console.log(<span class="code-string">\`Ad filled CPM: \${ad.cpm}, DSP: \${ad.dspId}\`</span>);
}).catch(err => {
  console.error(<span class="code-string">'Mediation error:'</span>, err);
});`
};

function initSDKTabs() {
  const tabBtns = document.querySelectorAll('.sdk-tab-btn');
  const codeContent = document.getElementById('sdk-code-content');

  if (tabBtns.length && codeContent) {
    tabBtns.forEach(btn => {
      btn.addEventListener('click', () => {
        // 去除所有 active
        tabBtns.forEach(b => b.classList.remove('active'));
        btn.classList.add('active');

        // 渲染对应语言
        const lang = btn.getAttribute('data-lang');
        codeContent.innerHTML = sdkSnippets[lang];
      });
    });
  }
}

// 4. 实时竞价模拟器引擎 (精修硬核日志，拒绝AI感官)
function initBiddingSimulator() {
  const startBtn = document.getElementById('start-sim-btn');
  const modeBtns = document.querySelectorAll('.sim-mode-btn');
  const indicator = document.getElementById('sim-indicator');
  const creativePanel = document.getElementById('sim-creative');
  const consoleBody = document.getElementById('sim-console-body');

  // 节点 DOM
  const nodeSdk = document.getElementById('node-sdk');
  const nodeCore = document.getElementById('node-core');
  const nodeDspA = document.getElementById('node-dsp-a');
  const nodeDspB = document.getElementById('node-dsp-b');
  const nodeDspC = document.getElementById('node-dsp-c');

  // SVG 动画线
  const pulseSdkCore = document.getElementById('pulse-sdk-core');
  const pulseCoreDspA = document.getElementById('pulse-core-dsp-a');
  const pulseCoreDspB = document.getElementById('pulse-core-dsp-b');
  const pulseCoreDspC = document.getElementById('pulse-core-dsp-c');

  let currentMode = 's2s'; // 's2s' 或 'waterfall'
  let isSimulating = false;

  // 模式切换
  modeBtns.forEach(btn => {
    btn.addEventListener('click', () => {
      if (isSimulating) return;
      modeBtns.forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      currentMode = btn.getAttribute('data-mode');
      clearConsole();
      addLog(`[SYS] Switch to [${currentMode === 's2s' ? 'S2S Bidding' : 'Waterfall'}] mode. Ready for simulation.`, 'info');
    });
  });

  // 日志助手
  function addLog(text, type = '') {
    const time = new Date().toLocaleTimeString('zh-CN', { hour12: false });
    const line = document.createElement('div');
    line.className = 'log-line';

    let spanClass = 'log-info';
    if (type === 'success') spanClass = 'log-success';
    if (type === 'warning') spanClass = 'log-warning';
    if (type === 'accent') spanClass = 'log-accent';

    line.innerHTML = `[${time}] <span class="${spanClass}">${text}</span>`;
    consoleBody.appendChild(line);
    consoleBody.scrollTop = consoleBody.scrollHeight;
  }

  function clearConsole() {
    consoleBody.innerHTML = '';
  }

  // 模拟过程
  async function runSimulation() {
    isSimulating = true;
    startBtn.disabled = true;
    startBtn.innerText = 'Bidding...';
    startBtn.style.opacity = '0.6';

    // 初始化节点样式
    resetSimulatorNodes();
    clearConsole();

    indicator.classList.add('active');

    // 步骤 1：SDK 发起广告加载
    addLog('[SDK] Request: loading interstitial placement "plc_home_inter" (width=1080, height=1920)...', 'accent');
    nodeSdk.style.borderColor = 'var(--primary)';
    nodeSdk.style.background = 'rgba(99, 102, 241, 0.05)';

    // 激活 SDK -> Core 的连接线
    pulseSdkCore.style.display = 'block';
    pulseSdkCore.style.animation = 'drawPath 0.8s forwards linear';

    await sleep(800);

    // 步骤 2：核心服务接收并做决策
    addLog('[ENGINE] Recv: incoming request from tenant "Default" (app_key: ab_prod_8f9c0e4).', 'info');
    nodeCore.classList.add('active');

    await sleep(600);

    if (currentMode === 's2s') {
      // --- S2S 模式 (高并发并行) ---
      addLog('[ENGINE] Match: placement strategy is Server-to-Server parallel bidding.', 'info');
      addLog('[ENGINE] S2S: dispatching parallel auctions to 3 configured DSPs...', 'info');

      // 激活 Core -> DSPs 的连接线
      pulseCoreDspA.style.display = 'block';
      pulseCoreDspB.style.display = 'block';
      pulseCoreDspC.style.display = 'block';

      pulseCoreDspA.style.animation = 'drawPath 0.6s forwards linear';
      pulseCoreDspB.style.animation = 'drawPath 0.6s forwards linear';
      pulseCoreDspC.style.animation = 'drawPath 0.6s forwards linear';

      await sleep(400);

      // 开启价格快速滚动模拟
      const rollA = startPriceRolling(nodeDspA.querySelector('.sim-node-price'), 0.5, 5.0);
      const rollB = startPriceRolling(nodeDspB.querySelector('.sim-node-price'), 1.0, 9.0);
      const rollC = startPriceRolling(nodeDspC.querySelector('.sim-node-price'), 0.5, 4.0);

      await sleep(1000);

      // 锁定价格
      const priceA = 4.20;
      const priceB = 8.65;
      const priceC = 3.50;

      clearInterval(rollA);
      clearInterval(rollB);
      clearInterval(rollC);

      lockPrice(nodeDspA, priceA);
      lockPrice(nodeDspB, priceB);
      lockPrice(nodeDspC, priceC);

      addLog(`[DSP-A] Bid response: 200 OK | bid_id: tx_b19fa2 | cpm: $${priceA.toFixed(2)} | latency: 18ms`, 'info');
      addLog(`[DSP-B] Bid response: 200 OK | bid_id: al_a7b8e9 | cpm: $${priceB.toFixed(2)} | latency: 32ms`, 'info');
      addLog(`[DSP-C] Bid response: 200 OK | bid_id: bd_c2f8a1 | cpm: $${priceC.toFixed(2)} | latency: 14ms`, 'info');

      await sleep(600);

      // 选定 Winner
      addLog(`[ENGINE] Auction: evaluation complete. Winner: DSP-B | CPM: $${priceB.toFixed(2)} USD | delta: +$${(priceB - priceA).toFixed(2)}`, 'success');
      nodeDspA.classList.add('loser');
      nodeDspC.classList.add('loser');
      nodeDspB.classList.add('winner');

      await sleep(500);

      // 回传给 SDK
      addLog('[ENGINE] Render: generating VAST 4.2 response with material type "video".', 'info');
      pulseCoreDspB.style.animation = 'drawPathReverse 0.6s forwards linear';
      await sleep(600);

      pulseSdkCore.style.animation = 'drawPathReverse 0.8s forwards linear';
      await sleep(800);

      // 渲染展示
      creativePanel.classList.add('loaded');
      creativePanel.innerHTML = `DSP-B (AppLovin) Video Creative Rendered | CPM: $${priceB.toFixed(2)}`;

      addLog(`[SDK] Render: VAST XML successfully parsed. Video creative loaded in 82ms.`, 'success');
      addLog(`[SDK] WinNotice: callback trigger sent to /api/v1/track?event=win&bid_id=al_a7b8e9`, 'success');

    } else {
      // --- Waterfall 瀑布流模式 (串行查询) ---
      addLog('[ENGINE] Match: placement strategy is Waterfall (sequential tier backup).', 'info');

      // 1. 请求 Tier 1 (DSP-A)
      addLog('[ENGINE] Waterfall: [Tier 1] querying DSP-A (tencent-adnetwork) | floor_price: 5.00 USD', 'info');
      pulseCoreDspA.style.display = 'block';
      pulseCoreDspA.style.animation = 'drawPath 0.8s forwards linear';
      await sleep(800);

      const rollA = startPriceRolling(nodeDspA.querySelector('.sim-node-price'), 0.5, 5.0);
      await sleep(800);
      clearInterval(rollA);

      addLog('[DSP-A] Bid response: 204 No Content | floor_price check failed | latency: 800ms (timeout)', 'warning');
      nodeDspA.classList.add('loser');
      lockPrice(nodeDspA, 0.0);

      await sleep(600);

      // 2. 请求 Tier 2 (DSP-B)
      addLog('[ENGINE] Waterfall: [Tier 2] querying DSP-B (applovin-exchange) | floor_price: 3.00 USD', 'info');
      pulseCoreDspB.style.display = 'block';
      pulseCoreDspB.style.animation = 'drawPath 0.8s forwards linear';
      await sleep(800);

      const rollB = startPriceRolling(nodeDspB.querySelector('.sim-node-price'), 1.0, 5.0);
      await sleep(800);
      clearInterval(rollB);

      const priceB = 4.50;
      lockPrice(nodeDspB, priceB);
      addLog(`[DSP-B] Bid response: 200 OK | bid_id: al_v19bf4 | cpm: $${priceB.toFixed(2)} | latency: 52ms`, 'success');
      nodeDspB.classList.add('winner');

      // DSP-C 被跳过
      addLog('[ENGINE] Waterfall: successfully filled at Tier 2. Skipping remaining lower tiers.', 'info');
      nodeDspC.classList.add('loser');

      await sleep(600);

      // 回传并渲染
      pulseCoreDspB.style.animation = 'drawPathReverse 0.6s forwards linear';
      await sleep(600);

      pulseSdkCore.style.animation = 'drawPathReverse 0.8s forwards linear';
      await sleep(800);

      creativePanel.classList.add('loaded');
      creativePanel.innerHTML = `DSP-B (AppLovin) Banner Creative Rendered | CPM: $${priceB.toFixed(2)}`;

      addLog(`[SDK] Render: Banner creative rendered. Total sequential latency accumulated: 285ms.`, 'warning');
    }

    // 恢复状态
    isSimulating = false;
    startBtn.disabled = false;
    startBtn.innerText = 'Run Auction Simulation';
    startBtn.style.opacity = '1';
    indicator.classList.remove('active');
  }

  function resetSimulatorNodes() {
    [nodeSdk, nodeCore, nodeDspA, nodeDspB, nodeDspC].forEach(node => {
      node.style.borderColor = '';
      node.style.background = '';
      node.className = 'sim-node';
      const priceEl = node.querySelector('.sim-node-price');
      if (priceEl) {
        priceEl.style.display = 'none';
        priceEl.innerText = '';
      }
    });

    // 清理连接线
    [pulseSdkCore, pulseCoreDspA, pulseCoreDspB, pulseCoreDspC].forEach(line => {
      line.style.display = 'none';
      line.style.animation = 'none';
    });

    creativePanel.className = 'sim-creative-panel';
    creativePanel.innerHTML = '准备就绪';
  }

  function startPriceRolling(element, min, max) {
    element.style.display = 'block';
    return setInterval(() => {
      const randomPrice = (Math.random() * (max - min) + min).toFixed(2);
      element.innerText = `$${randomPrice}`;
    }, 80);
  }

  function lockPrice(node, price) {
    const el = node.querySelector('.sim-node-price');
    if (el) {
      el.style.display = 'block';
      if (price > 0) {
        el.innerText = `$${price.toFixed(2)}`;
      } else {
        el.innerText = 'NO FILL';
        el.style.color = 'var(--text-muted)';
      }
    }
  }

  function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  startBtn.addEventListener('click', () => {
    if (isSimulating) return;
    runSimulation();
  });
}

// 5. 复制 Docker 部署指令
function initDockerCopy() {
  const copyBtn = document.getElementById('copy-docker-btn');
  const codeText = document.getElementById('docker-code-text');

  if (copyBtn && codeText) {
    copyBtn.addEventListener('click', () => {
      const code = codeText.innerText.replace('$', '').trim();
      navigator.clipboard.writeText(code).then(() => {
        const origIcon = copyBtn.innerHTML;
        copyBtn.innerHTML = '✓ Copied';
        copyBtn.style.color = 'var(--success)';
        setTimeout(() => {
          copyBtn.innerHTML = origIcon;
          copyBtn.style.color = '';
        }, 2000);
      });
    });
  }
}

// 6. 苹果 Pro 级全屏开发者文档控制台逻辑
function initDocsModal() {
  const openBtn = document.getElementById('open-docs-btn');
  const closeBtn = document.getElementById('close-docs-btn');
  const modal = document.getElementById('docs-modal');
  const menuItems = document.querySelectorAll('.docs-menu-item');
  const docPages = document.querySelectorAll('.doc-page');

  if (!openBtn || !closeBtn || !modal) return;

  // 打开 Drawer
  openBtn.addEventListener('click', (e) => {
    e.preventDefault();
    modal.classList.add('active');
    document.body.style.overflow = 'hidden'; // 锁死背景滚动，保障苹果原生高阶体验
  });

  // 关闭 Drawer
  const closeModal = () => {
    modal.classList.remove('active');
    document.body.style.overflow = '';
  };

  closeBtn.addEventListener('click', closeModal);

  // 点击外部遮罩区关闭
  modal.addEventListener('click', (e) => {
    if (e.target === modal) {
      closeModal();
    }
  });

  // 左侧菜单 Tab 切换逻辑
  menuItems.forEach(item => {
    item.addEventListener('click', () => {
      // 移除所有 active
      menuItems.forEach(i => i.classList.remove('active'));
      docPages.forEach(p => p.classList.remove('active'));

      // 激活当前
      item.classList.add('active');
      const targetDoc = item.getAttribute('data-doc');
      const targetPage = document.getElementById(`doc-${targetDoc}`);
      if (targetPage) {
        targetPage.classList.add('active');
      }
    });
  });
}
