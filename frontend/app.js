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
});
