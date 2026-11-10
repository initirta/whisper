import Link from "next/link";

const steps = [
  ["01", "Buat link pribadimu", "Bagikan link Whisper ke teman-temanmu."],
  ["02", "Terima pesan anonim", "Siapa pun bisa menulis tanpa membuat akun."],
  ["03", "Lanjutkan percakapan", "Balas pesan secara privat, kapan saja."],
];

export default function Home() {
  return (
    <main className="landing-shell">
      <header className="topbar page-width">
        <Link className="brand" href="/" aria-label="Whisper beranda"><span className="brand-mark">w</span> whisper</Link>
        <nav><Link href="/admin/login">Masuk</Link><Link className="nav-cta" href="/admin/login">Mulai sekarang <span>↗</span></Link></nav>
      </header>
      <section className="hero page-width">
        <div className="hero-copy">
          <div className="eyebrow"><span className="status-dot" /> RUANG AMAN UNTUK BICARA</div>
          <h1>Hal yang sulit<br />diucapkan, <em>bisa ditulis.</em></h1>
          <p className="hero-lede">Terima pesan jujur dari orang-orang di sekitarmu. Tanpa nama, tanpa tekanan—cukup percakapan yang berarti.</p>
          <div className="hero-actions"><Link href="/admin/login" className="button button-primary">Buat ruang Whisper <span>→</span></Link><a href="#cara-kerja" className="text-link">Lihat cara kerja <span>↓</span></a></div>
          <div className="privacy-note"><span className="lock-icon">◇</span><span>Pesan bersifat privat. Identitas pengirim tidak ditampilkan.</span></div>
        </div>
        <div className="hero-art" aria-label="Ilustrasi percakapan Whisper">
          <div className="orbit orbit-one" /><div className="orbit orbit-two" />
          <div className="spark spark-one">✳</div><div className="spark spark-two">✦</div>
          <div className="message-card card-back"><div className="avatar avatar-peach">R</div><div><span className="mini-label">PESAN ANONIM</span><p>Aku selalu kagum sama caramu<br />tetap jadi diri sendiri.</p></div><span className="card-time">baru saja</span></div>
          <div className="message-card card-front"><div className="card-head"><div className="avatar avatar-lilac">?</div><div><strong>Seseorang ingin bicara</strong><span>pesan baru · anonim</span></div><span className="online-dot" /></div><div className="sample-bubble">Makasih sudah jadi teman yang selalu ada. ✨</div><div className="sample-reply">Selalu ada buat kamu. 💜</div><div className="card-foot"><span>percakapan privat</span><span className="pulse-bars"><i /><i /><i /></span></div></div>
          <div className="floating-note"><span>♡</span> ruang tanpa penilaian</div>
        </div>
      </section>
      <section className="trust-strip"><div className="page-width trust-inner"><span>UNTUK CERITA YANG LEBIH JUJUR</span><i /> <span>ANONIM</span><i /> <span>PRIVAT</span><i /> <span>SELALU TERHUBUNG</span></div></section>
      <section className="steps-section page-width" id="cara-kerja"><div className="section-heading"><span className="eyebrow">SEDERHANA, SEPERTI SEHARUSNYA</span><h2>Mulai percakapan yang<br /><em>berarti.</em></h2></div><div className="steps-grid">{steps.map(([number,title,description])=><article className="step-card" key={number}><span className="step-number">{number}</span><div className="step-line" /><h3>{title}</h3><p>{description}</p></article>)}</div></section>
      <footer className="page-width footer"><Link className="brand" href="/"><span className="brand-mark">w</span> whisper</Link><span>Anonim bukan berarti tanpa batas. Jaga percakapan tetap baik.</span><Link href="/admin/login">Masuk ke ruangmu ↗</Link></footer>
    </main>
  );
}
