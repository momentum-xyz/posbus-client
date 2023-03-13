// esbuild wasm import
declare module "*.wasm" {
  const content: string;
  export default content;
}
