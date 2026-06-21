document.addEventListener('DOMContentLoaded', () => {
    const runBtn = document.getElementById('run-btn');
    const timeVal = document.getElementById('time-val');
    const planJson = document.getElementById('plan-json');

    runBtn.addEventListener('click', async () => {
        // UIリセット
        runBtn.textContent = 'Running...';
        runBtn.disabled = true;
        timeVal.textContent = '...';
        planJson.textContent = 'データベースにクエリを送信し、実行計画を取得中...';

        try {
            // バックエンドAPIをコール
            const response = await fetch('http://localhost:8080/api/query');
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();

            if (data.error) {
                throw new Error(data.error);
            }

            // 結果の反映
            timeVal.textContent = data.executionTimeMs;
            
            // JSONのハイライト（簡易的に表示）
            planJson.textContent = JSON.stringify(data.plan, null, 2);

        } catch (error) {
            console.error('Error fetching query plan:', error);
            timeVal.textContent = 'ERR';
            planJson.textContent = `エラーが発生しました: ${error.message}\nバックエンド(Go)サーバーとDB(Podman)が起動しているか確認してください。`;
        } finally {
            runBtn.textContent = 'Run Query & Get Plan';
            runBtn.disabled = false;
        }
    });

    const clearBtn = document.getElementById('clear-btn');
    clearBtn.addEventListener('click', async () => {
        const originalText = clearBtn.textContent;
        clearBtn.textContent = 'Restarting DB...';
        clearBtn.disabled = true;
        runBtn.disabled = true;
        
        try {
            const response = await fetch('http://localhost:8080/api/clearcache', { method: 'POST' });
            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
            const data = await response.json();
            if (data.error) throw new Error(data.error);
            
            alert('DBを再起動し、キャッシュ（共有バッファ）をクリアしました！');
        } catch (error) {
            console.error('Error clearing cache:', error);
            alert(`キャッシュのクリアに失敗しました: ${error.message}`);
        } finally {
            clearBtn.textContent = originalText;
            clearBtn.disabled = false;
            runBtn.disabled = false;
        }
    });

    // === タブ切り替え処理 ===
    const tabBtns = document.querySelectorAll('.tab-btn');
    const tabContents = document.querySelectorAll('.tab-content');

    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            // タブボタンのアクティブ切り替え
            tabBtns.forEach(b => b.classList.remove('active'));
            btn.classList.add('active');

            // コンテンツのアクティブ切り替え
            const targetId = btn.getAttribute('data-target');
            tabContents.forEach(content => {
                if (content.id === targetId) {
                    content.classList.add('active');
                } else {
                    content.classList.remove('active');
                }
            });
        });
    });

    // === Demo 2 処理 ===
    const runBtn2 = document.getElementById('run-btn2');
    const timeVal2 = document.getElementById('time-val2');
    const planJson2 = document.getElementById('plan-json2');

    runBtn2.addEventListener('click', async () => {
        runBtn2.textContent = 'Running...';
        runBtn2.disabled = true;
        timeVal2.textContent = '...';
        planJson2.textContent = 'MySQLにクエリを送信し、実行計画を取得中...';

        try {
            const response = await fetch('http://localhost:8080/api/demo2/query');
            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
            const data = await response.json();

            if (data.error) throw new Error(data.error);

            timeVal2.textContent = data.executionTimeMs;
            planJson2.textContent = JSON.stringify(data.plan, null, 2);
        } catch (error) {
            console.error('Error:', error);
            timeVal2.textContent = 'ERR';
            planJson2.textContent = `エラー: ${error.message}`;
        } finally {
            runBtn2.textContent = 'Run Query & Get Plan';
            runBtn2.disabled = false;
        }
    });

    const clearBtn2 = document.getElementById('clear-btn2');
    clearBtn2.addEventListener('click', async () => {
        const originalText = clearBtn2.textContent;
        clearBtn2.textContent = 'Restarting DB...';
        clearBtn2.disabled = true;
        runBtn2.disabled = true;
        
        try {
            const response = await fetch('http://localhost:8080/api/demo2/clearcache', { method: 'POST' });
            if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
            const data = await response.json();
            if (data.error) throw new Error(data.error);
            
            alert('MySQLコンテナを再起動し、キャッシュをクリアしました！');
        } catch (error) {
            console.error('Error clearing cache:', error);
            alert(`キャッシュクリア失敗: ${error.message}`);
        } finally {
            clearBtn2.textContent = originalText;
            clearBtn2.disabled = false;
            runBtn2.disabled = false;
        }
    });
});
