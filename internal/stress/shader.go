package stress

// ALU: 大量 FMA, 写回防止编译器优化掉.
const shaderALU = `
@group(0) @binding(0) var<storage, read_write> buf: array<f32>;

@compute @workgroup_size(256)
fn main(@builtin(global_invocation_id) gid: vec3<u32>) {
    let n = arrayLength(&buf);
    var acc = f32(gid.x % n) * 0.001;
    for (var i = 0u; i < 4096u; i++) {
        acc = fma(acc, 1.0001, 0.9999);
    }
    buf[gid.x % n] = acc;
}
`

// MEM: 跨越 L2 cache 的大 buffer 读写.
const shaderMem = `
@group(0) @binding(0) var<storage, read_write> buf: array<f32>;

@compute @workgroup_size(256)
fn main(@builtin(global_invocation_id) gid: vec3<u32>) {
    let n = arrayLength(&buf);
    let i = gid.x % n;
    let j = (gid.x + n / 2u) % n;
    buf[j] = buf[i] * 2.0 + 1.0;
}
`
