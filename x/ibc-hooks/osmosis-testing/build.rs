extern crate core;

use std::{env, path::PathBuf, process::Command};

fn main() {
    let manifest_dir = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    let prebuilt_lib_dir = manifest_dir.join("libosmosistesting").join("artifacts");

    let lib_name = "osmosistesting";

    let out_dir = PathBuf::from(env::var("OUT_DIR").unwrap());

    let header = if std::env::var("DOCS_RS").is_ok() {
        manifest_dir
            .join("libosmosistesting")
            .join("artifacts")
            .join("libosmosistesting.docrs.h")
    } else {
        out_dir.join(format!("lib{}.h", lib_name))
    };
    // rerun when go code is updated
    println!("cargo:rerun-if-changed=./libosmosistesting");

    let lib_filename = if cfg!(target_os = "macos") {
        format!("lib{}.{}", lib_name, "dylib")
    } else if cfg!(target_os = "linux") {
        format!("lib{}.{}", lib_name, "so")
    } else if cfg!(target_os = "windows") {
        // untested
        format!("{}.{}", lib_name, "dll")
    } else {
        panic!("Unsupported architecture");
    };

    let lib_filename = lib_filename.as_str();

    if env::var("PREBUILD_LIB") == Ok("1".to_string()) {
        build_libosmosistesting(prebuilt_lib_dir.join(lib_filename));
    }

    let out_dir_lib_path = out_dir.join(lib_filename);
    build_libosmosistesting(out_dir_lib_path);

    // define lib name
    println!(
        "cargo:rustc-link-search=native={}",
        out_dir.to_str().unwrap()
    );

    // disable linking if docrs
    if std::env::var("DOCS_RS").is_err() {
        println!("cargo:rustc-link-lib=dylib={}", lib_name);
    }

    // The bindgen::Builder is the main entry point
    // to bindgen, and lets you build up options for
    // the resulting bindings.
    let bindings = bindgen::Builder::default()
        // The input header we would like to generate
        // bindings for.
        .header(header.to_str().unwrap())
        // Tell cargo to invalidate the built crate whenever any of the
        // included header files changed.
        .parse_callbacks(Box::new(bindgen::CargoCallbacks))
        // Finish the builder and generate the bindings.
        .generate()
        // Unwrap the Result and panic on failure.
        .expect("Unable to generate bindings");

    // Write the bindings to the $OUT_DIR/bindings.rs file.
    let out_path = PathBuf::from(env::var("OUT_DIR").unwrap());
    bindings
        .write_to_file(out_path.join("bindings.rs"))
        .expect("Couldn't write bindings!");
}

fn build_libosmosistesting(out: PathBuf) {
    // skip if doc_rs build
    if std::env::var("DOCS_RS").is_ok() {
        return;
    }
    let manifest_dir = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    let exit_status = Command::new("go")
        .current_dir(manifest_dir.join("libosmosistesting"))
        .arg("build")
        .arg("-buildmode=c-shared")
        .arg("-o")
        .arg(out)
        .arg("main.go")
        .spawn()
        .unwrap()
        .wait()
        .unwrap();

    if !exit_status.success() {
        panic!("failed to build go code");
    }
}
