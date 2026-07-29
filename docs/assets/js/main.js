/* ==========================================================================
   CatNet TUI — Vanilla JavaScript for Tabs and Clipboard
   No third-party libraries or external CDNs
   ========================================================================== */

document.addEventListener('DOMContentLoaded', () => {
  // Tab Switching Functionality
  const tabButtons = document.querySelectorAll('.tab-button');
  const tabPanels = document.querySelectorAll('.tab-panel');

  tabButtons.forEach(button => {
    button.addEventListener('click', () => {
      const targetTabId = button.getAttribute('data-tab');

      // Remove active class from all buttons and panels
      tabButtons.forEach(btn => btn.classList.remove('active'));
      tabPanels.forEach(panel => panel.classList.remove('active'));

      // Activate selected button and target panel
      button.classList.add('active');
      const targetPanel = document.getElementById(targetTabId);
      if (targetPanel) {
        targetPanel.classList.add('active');
      }
    });
  });

  // Code Block Copy to Clipboard
  const copyButtons = document.querySelectorAll('.copy-btn');

  copyButtons.forEach(button => {
    button.addEventListener('click', async () => {
      const targetId = button.getAttribute('data-copy-target');
      let textToCopy = '';

      if (targetId) {
        const targetElement = document.getElementById(targetId);
        if (targetElement) {
          textToCopy = targetElement.textContent.trim();
        }
      } else {
        const wrapper = button.closest('.code-block-wrapper');
        if (wrapper) {
          const codeEl = wrapper.querySelector('code, pre');
          if (codeEl) {
            textToCopy = codeEl.textContent.trim();
          }
        }
      }

      if (!textToCopy) return;

      try {
        await navigator.clipboard.writeText(textToCopy);
        const originalText = button.innerHTML;
        button.classList.add('copied');
        button.innerHTML = '✓ Copied!';

        setTimeout(() => {
          button.classList.remove('copied');
          button.innerHTML = originalText;
        }, 2000);
      } catch (err) {
        console.error('Failed to copy text: ', err);
      }
    });
  });
});
