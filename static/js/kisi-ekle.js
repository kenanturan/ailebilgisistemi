function ekleKisi() {
    console.log('ekleKisi fonksiyonu çağrıldı');
    
    const kisi = {
        ad: document.getElementById('ad').value,
        soyad: document.getElementById('soyad').value,
        tc: document.getElementById('tc').value,
        cepTelefonu: document.getElementById('cepTelefonu').value,
        anneAdi: document.getElementById('anneAdi').value,
        babaAdi: document.getElementById('babaAdi').value,
        esId: document.getElementById('esId').value,
        cinsiyet: document.getElementById('cinsiyet').value,
        hakkinda: document.getElementById('hakkinda').value,
        fotograf: document.getElementById('fotografOnizleme')?.src || ''
    };

    // Zorunlu alan kontrolü
    if (!kisi.ad || !kisi.soyad || !kisi.tc || !kisi.cinsiyet) {
        alert('Lütfen zorunlu alanları doldurun (Ad, Soyad, TC, Cinsiyet)');
        return;
    }

    // TC Kimlik kontrolü
    if (kisi.tc.length !== 11) {
        alert('TC Kimlik numarası 11 haneli olmalıdır');
        return;
    }

    console.log('Gönderilecek kişi bilgileri:', kisi);

    fetch('/api/person/create', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(kisi)
    })
    .then(response => {
        console.log('API yanıtı:', response);
        if (!response.ok) {
            return response.text().then(text => {
                throw new Error(text || 'Kişi eklenirken bir hata oluştu');
            });
        }
        return response.json();
    })
    .then(data => {
        console.log('Başarılı yanıt:', data);
        alert('Kişi başarıyla eklendi');
        window.location.href = '/kisi-listesi';
    })
    .catch(error => {
        console.error('Hata:', error);
        alert('Hata: ' + error.message);
    });
}

// Fotoğraf seçildiğinde önizleme göster
document.getElementById('fotograf')?.addEventListener('change', function(event) {
    const file = event.target.files[0];
    if (file) {
        const reader = new FileReader();
        reader.onload = function(e) {
            const img = document.getElementById('fotografOnizleme');
            if (img) {
                img.src = e.target.result;
                img.style.display = 'block';
            }
        };
        reader.readAsDataURL(file);
    }
});

// Anne-baba listelerini doldur
fetch('/api/people')
    .then(response => response.json())
    .then(data => {
        const anneSelect = document.getElementById('anneAdi');
        const babaSelect = document.getElementById('babaAdi');
        const esSelect = document.getElementById('esId');
        
        if (anneSelect && babaSelect && esSelect) {
            data.forEach(kisi => {
                const option = document.createElement('option');
                option.value = kisi.id;
                option.textContent = kisi.ad + ' ' + kisi.soyad;
                
                if (kisi.cinsiyet === 'Kadın') {
                    anneSelect.appendChild(option.cloneNode(true));
                }
                if (kisi.cinsiyet === 'Erkek') {
                    babaSelect.appendChild(option.cloneNode(true));
                }
                esSelect.appendChild(option.cloneNode(true));
            });
        }
    })
    .catch(error => console.error('Ebeveyn/Eş listesi yüklenirken hata:', error));