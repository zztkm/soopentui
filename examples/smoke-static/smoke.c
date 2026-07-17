/* Minimal static-link smoke test against libopentui.a */
#include <stdint.h>
#include <stdio.h>

typedef uint32_t NativeHandle;

NativeHandle createRenderer(
    uint32_t width,
    uint32_t height,
    uint8_t buffered_destination_kind,
    uint8_t remote_mode_value,
    void *feed_ptr);
void destroyRenderer(NativeHandle renderer);

int main(void) {
  /* bufferedDestinationKind=1 => memory; remoteModeValue=1 => local */
  NativeHandle r = createRenderer(80, 24, 1, 1, NULL);
  printf("renderer=%u\n", (unsigned)r);
  if (r == 0) {
    return 1;
  }
  destroyRenderer(r);
  return 0;
}
