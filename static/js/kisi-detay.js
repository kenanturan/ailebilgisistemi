document.addEventListener('DOMContentLoaded', function() {
    // Kişi listesinden fotoğrafı al ve göster
    const kisiId = window.location.pathname.split('/').pop();
    fetch('/api/people')
        .then(response => response.json())
        .then(data => {
            const kisi = data.find(k => k.id === kisiId);
            if (kisi) {
                const img = document.getElementById('fotografOnizleme');
                img.src = kisi.fotograf || DEFAULT_PHOTO;
                img.style.display = 'block';
            }
        });

    // Form submit olayını dinle
    document.getElementById('kisiForm').addEventListener('submit', function(e) {
        e.preventDefault();
        kisiGuncelle();
    });

    // Fotoğraf değişikliğini dinle
    document.getElementById('fotograf').addEventListener('change', onFotoSecildi);

    // Anne-baba listelerini doldur
    ebeveynListesiniDoldur();

    // Mevcut cinsiyet değerini seç
    const cinsiyet = document.getElementById('cinsiyet').getAttribute('data-value');
    if (cinsiyet) {
        document.getElementById('cinsiyet').value = cinsiyet;
    }

    esListesiniDoldur();
});

// Varsayılan fotoğraf
const DEFAULT_PHOTO = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=';

function onFotoSecildi(event) {
    const file = event.target.files[0];
    if (file) {
        const reader = new FileReader();
        reader.onload = function(e) {
            const img = document.getElementById('fotografOnizleme');
            img.src = e.target.result || DEFAULT_PHOTO;
            img.style.display = 'block';
        };
        reader.readAsDataURL(file);
    }
}

function kisiGuncelle() {
    const kisiId = window.location.pathname.split('/').pop();
    const img = document.getElementById('fotografOnizleme');
    const kisi = {
        id: kisiId,
        ad: document.getElementById('ad').value,
        soyad: document.getElementById('soyad').value,
        tc: document.getElementById('tc').value,
        cepTelefonu: document.getElementById('cepTelefonu').value,
        anneAdi: document.getElementById('anneAdi').value,
        babaAdi: document.getElementById('babaAdi').value,
        esId: document.getElementById('esId').value,
        cinsiyet: document.getElementById('cinsiyet').value,
        hakkinda: document.getElementById('hakkinda').value,
        fotograf: img.src || DEFAULT_PHOTO
    };

    console.log('Sunucuya gönderilen fotoğraf verisi var mı?:', !!kisi.fotograf);

    fetch('/api/person/update', {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(kisi)
    })
    .then(response => {
        if (response.ok) {
            alert('Kişi başarıyla güncellendi');
            window.location.reload();
        } else {
            response.text().then(error => alert('Hata: ' + error));
        }
    })
    .catch(error => alert('Bir hata oluştu: ' + error));
}

function silKisi(id) {
    if (confirm('Bu kişiyi silmek istediğinizden emin misiniz?')) {
        fetch(`/api/person/delete?id=${id}`, {
            method: 'DELETE'
        })
        .then(response => {
            if (response.ok) {
                alert('Kişi başarıyla silindi');
                // Silme işleminden sonra kişi listesine yönlendir
                window.location.href = '/kisi-listesi';
            } else {
                alert('Silme işlemi başarısız oldu');
            }
        })
        .catch(error => alert('Bir hata oluştu: ' + error));
    }
}

function ebeveynListesiniDoldur() {
    fetch('/api/people')
        .then(response => response.json())
        .then(data => {
            const anneSelect = document.getElementById('anneAdi');
            const babaSelect = document.getElementById('babaAdi');
            
            const mevcutAnne = anneSelect.getAttribute('data-value');
            const mevcutBaba = babaSelect.getAttribute('data-value');
            
            data.forEach(kisi => {
                if (kisi.cinsiyet === 'Kadın') {
                    const option = document.createElement('option');
                    option.value = kisi.id;
                    option.textContent = kisi.ad + ' ' + kisi.soyad;
                    option.selected = kisi.id === mevcutAnne;
                    anneSelect.appendChild(option);
                }
                if (kisi.cinsiyet === 'Erkek') {
                    const option = document.createElement('option');
                    option.value = kisi.id;
                    option.textContent = kisi.ad + ' ' + kisi.soyad;
                    option.selected = kisi.id === mevcutBaba;
                    babaSelect.appendChild(option);
                }
            });
        });
}

function esListesiniDoldur() {
    fetch('/api/people')
        .then(response => response.json())
        .then(data => {
            const esSelect = document.getElementById('esId');
            const mevcutEs = esSelect.getAttribute('data-value');
            
            data.forEach(kisi => {
                // Kendisi hariç diğer kişileri listele
                if (kisi.id !== window.location.pathname.split('/').pop()) {
                    const option = document.createElement('option');
                    option.value = kisi.id;
                    option.textContent = kisi.ad + ' ' + kisi.soyad;
                    option.selected = kisi.id === mevcutEs;
                    esSelect.appendChild(option);
                }
            });
        });
}

function buyukFotografGoster(src) {
    const modal = document.getElementById('fotoModal');
    const modalImg = document.getElementById('buyukFotograf');
    modal.style.display = "block";
    modalImg.src = src;
}

function modalKapat() {
    document.getElementById('fotoModal').style.display = "none";
}

// Modal dışına tıklandığında kapanması için
window.onclick = function(event) {
    const modal = document.getElementById('fotoModal');
    if (event.target == modal) {
        modal.style.display = "none";
    }
}