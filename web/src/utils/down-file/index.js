export default function (content, name, type) {
  let blob;

  // 指定下载的文件类型
  switch (type) {
    case 'json':
      blob = new Blob([content]);
      break;
    default:
      blob = content;
  }
  if (window.navigator.msSaveOrOpenBlob) {
    navigator.msSaveBlob(blob, `${name}.${type}`);
  } else {
    const link = document.createElement('a');

    link.href = window.URL.createObjectURL(blob);
    link.download = `${name}.${type}`;
    link.click();
    window.URL.revokeObjectURL(link.href);
  }
}
