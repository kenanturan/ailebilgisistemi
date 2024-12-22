// Varsayılan fotoğraf
const DEFAULT_PHOTO = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=';

document.addEventListener('DOMContentLoaded', function() {
    console.log('Sayfa yüklendi');
    kisileriGetir();
    
    // Arama fonksiyonu
    const searchInput = document.getElementById('searchInput');
    if (searchInput) {
        searchInput.addEventListener('input', function(e) {
            const searchText = e.target.value.toLowerCase();
            const rows = document.querySelectorAll('#kisiListesi tbody tr');
            
            rows.forEach(row => {
                const text = row.textContent.toLowerCase();
                row.style.display = text.includes(searchText) ? '' : 'none';
            });
        });
    }
});

function kisileriGetir() {
    console.log('Kişiler getiriliyor...');
    
    fetch('/api/people')
        .then(response => response.json())
        .then(data => {
            const tbody = document.querySelector('#kisiListesi tbody');
            tbody.innerHTML = '';
            
            if (data && data.length > 0) {
                // Eş bilgilerini bulmak için kişileri bir Map'te tutalım
                const kisilerMap = new Map(data.map(k => [k.id, k]));
                
                data.forEach(kisi => {
                    // Eşin adını ve soyadını bulalım
                    let esAdSoyad = '';
                    if (kisi.esId) {
                        const es = kisilerMap.get(kisi.esId);
                        if (es) {
                            esAdSoyad = `${es.ad} ${es.soyad}`;
                        }
                    }

                    const row = document.createElement('tr');
                    row.innerHTML = `
                        <td><img src="${kisi.fotograf || DEFAULT_PHOTO}" class="profile-img"></td>
                        <td>${kisi.ad || ''}</td>
                        <td>${kisi.soyad || ''}</td>
                        <td>${kisi.tc || ''}</td>
                        <td>${kisi.cepTelefonu || ''}</td>
                        <td>${kisi.anneAdSoyad || ''}</td>
                        <td>${kisi.babaAdSoyad || ''}</td>
                        <td>${esAdSoyad || ''}</td>
                        <td>${kisi.cinsiyet || ''}</td>
                        <td>${kisi.hakkinda || ''}</td>
                    `;
                    row.style.cursor = 'pointer';
                    row.onclick = () => window.location.href = `/kisi/${kisi.id}`;
                    tbody.appendChild(row);
                });
            } else {
                tbody.innerHTML = `
                    <tr>
                        <td colspan="10" style="text-align: center;">Henüz kişi eklenmemiş</td>
                    </tr>
                `;
            }
        })
        .catch(error => {
            console.error('Veri getirme hatası:', error);
            alert('Kişiler getirilirken bir hata oluştu: ' + error.message);
        });
}

function silKisi(id) {
    if (confirm('Bu kişiyi silmek istediğinizden emin misiniz?')) {
        fetch(`/api/person/delete?id=${id}`, {
            method: 'DELETE'
        })
        .then(response => {
            if (response.ok) {
                alert('Kişi başarıyla silindi');
                kisileriGetir();
            } else {
                alert('Silme işlemi başarısız oldu');
            }
        })
        .catch(error => alert('Bir hata oluştu: ' + error));
    }
}

function duzenleKisi(id) {
    // Kişinin bilgilerini al
    fetch(`/api/people`)
        .then(response => response.json())
        .then(data => {
            const kisi = data.find(k => k.id === id);
            if (kisi) {
                // Modal dialog oluştur
                const modal = document.createElement('div');
                modal.className = 'modal';
                modal.innerHTML = `
                    <div class="modal-content">
                        <h3>Kişi Düzenle</h3>
                        <form id="duzenleForm">
                            <div class="form-group">
                                <label>Ad:</label>
                                <input type="text" id="editAd" value="${kisi.ad}" required>
                            </div>
                            <div class="form-group">
                                <label>Soyad:</label>
                                <input type="text" id="editSoyad" value="${kisi.soyad}" required>
                            </div>
                            <div class="form-group">
                                <label>TC:</label>
                                <input type="text" id="editTC" value="${kisi.tc}" required pattern="[0-9]{11}">
                            </div>
                            <div class="form-group">
                                <label>Telefon:</label>
                                <input type="tel" id="editTelefon" value="${kisi.cepTelefonu || ''}" pattern="[0-9]{10}">
                            </div>
                            <div class="form-group">
                                <label>Anne:</label>
                                <select id="editAnne">
                                    <option value="">Seçiniz...</option>
                                </select>
                            </div>
                            <div class="form-group">
                                <label>Baba:</label>
                                <select id="editBaba">
                                    <option value="">Seçiniz...</option>
                                </select>
                            </div>
                            <div class="form-group">
                                <label>Cinsiyet:</label>
                                <select id="editCinsiyet" required>
                                    <option value="">Seçiniz...</option>
                                    <option value="Kadın">Kadın</option>
                                    <option value="Erkek">Erkek</option>
                                </select>
                            </div>
                            <div class="form-group">
                                <label>Hakkında:</label>
                                <textarea id="editHakkinda">${kisi.hakkinda || ''}</textarea>
                            </div>
                            <div class="form-group">
                                <label>Fotoğraf:</label>
                                <input type="file" id="editFotograf" accept="image/*">
                                <img id="editFotografOnizleme" src="${kisi.fotograf || ''}" style="max-width: 200px; margin-top: 10px;">
                            </div>
                            <div class="button-group">
                                <button type="submit" class="save-btn">Kaydet</button>
                                <button type="button" class="cancel-btn" onclick="modalKapat()">İptal</button>
                            </div>
                        </form>
                    </div>
                `;
                document.body.appendChild(modal);

                // Anne-baba listelerini doldur
                ebeveynListesiniDoldur('editAnne', 'editBaba', kisi.anneAdi, kisi.babaAdi);

                // Cinsiyet seç
                document.getElementById('editCinsiyet').value = kisi.cinsiyet;

                // Form submit olayını dinle
                document.getElementById('duzenleForm').addEventListener('submit', function(e) {
                    e.preventDefault();
                    guncelle(id);
                });

                // Fotoğraf değişikliğini dinle
                document.getElementById('editFotograf').addEventListener('change', function(e) {
                    const file = e.target.files[0];
                    if (file) {
                        const reader = new FileReader();
                        reader.onload = function(e) {
                            document.getElementById('editFotografOnizleme').src = e.target.result;
                        };
                        reader.readAsDataURL(file);
                    }
                });
            }
        });
}

function modalKapat() {
    const modal = document.querySelector('.modal');
    if (modal) {
        modal.remove();
    }
}

function guncelle(id) {
    const guncelKisi = {
        id: id,
        ad: document.getElementById('editAd').value,
        soyad: document.getElementById('editSoyad').value,
        tc: document.getElementById('editTC').value,
        cepTelefonu: document.getElementById('editTelefon').value,
        anneAdi: document.getElementById('editAnne').value,
        babaAdi: document.getElementById('editBaba').value,
        cinsiyet: document.getElementById('editCinsiyet').value,
        hakkinda: document.getElementById('editHakkinda').value,
        fotograf: document.getElementById('editFotografOnizleme').src
    };

    fetch('/api/person/update', {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(guncelKisi)
    })
    .then(response => {
        if (response.ok) {
            alert('Kişi başarıyla güncellendi');
            modalKapat();
            kisileriGetir();
        } else {
            response.text().then(error => alert('Hata: ' + error));
        }
    })
    .catch(error => alert('Bir hata oluştu: ' + error));
} 