export default function preset(context, opts = {}) {
  return {
    themes: [['docusaurus-theme-openapi-docs', opts.theme]],
    plugins: [
      ['@docusaurus/plugin-content-docs', { ...opts.docs1, id: 'docs1' }],
      ['@docusaurus/plugin-content-docs', { ...opts.docs2, id: 'docs2' }],
      ['@docusaurus/plugin-content-docs', { ...opts.docs3, id: 'docs3' }],
    ],
  };
}
